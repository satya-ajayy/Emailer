package kafka

import (
	// Go Internal Packages
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	// Local Packages
	models "emailer/models"

	// External Packages
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kprom"
	"go.uber.org/zap"
)

type ConsumerConfig struct {
	Brokers        []string
	Name           string
	Topic          string
	RecordsPerPoll int
}

type Consumer struct {
	Client    *kgo.Client
	Config    *ConsumerConfig
	Processor MailProcessor
	Logger    *zap.Logger
}

type MailProcessor interface {
	ProcessRecords(ctx context.Context, records []models.Record) error
}

// NewConsumer creates a new consumer to consume transactions topic
// (PS: Must call Poll to start consuming the records)
func NewConsumer(conf *ConsumerConfig, logger *zap.Logger, processor MailProcessor, metrics *kprom.Metrics) (*Consumer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(conf.Brokers...), // Connects to Kafka brokers
		kgo.ConsumerGroup(conf.Name),     // Specifies the consumer group
		kgo.ConsumeTopics(conf.Topic),    // Specifies a single topic to consume
		kgo.WithHooks(metrics),           // Attaches monitoring hooks
		kgo.DisableAutoCommit(),          // Disables auto-commit
		kgo.BlockRebalanceOnPoll(),       // Blocks rebalancing until the poll loop is running
	}

	client, err := kgo.NewClient(opts...)
	if err != nil || client == nil {
		return nil, err
	}

	return &Consumer{
		Client:    client,
		Config:    conf,
		Processor: processor,
		Logger:    logger,
	}, nil
}

// Poll polls for records from the Kafka broker.
func (c *Consumer) Poll(ctx context.Context) error {
	defer c.Client.Close()

	for {
		// Check if the context is canceled before polling
		if ctx.Err() != nil {
			c.Logger.Warn("polling stopped: context canceled")
			return ctx.Err() // Exit gracefully
		}

		c.Logger.Info(fmt.Sprintf("%s: polling for records", c.Config.Name))
		fetches := c.Client.PollRecords(ctx, c.Config.RecordsPerPoll)

		// Handle client shutdown
		if fetches.IsClientClosed() {
			return errors.New("kafka client closed")
		}

		// Handle context cancellation explicitly
		if errors.Is(fetches.Err0(), context.Canceled) {
			return errors.New("context got canceled")
		}

		// Preallocate records slice
		records := make([]models.Record, len(fetches.Records()))
		for idx, record := range fetches.Records() {
			records[idx] = models.Record{
				Key:   record.Key,
				Value: record.Value,
				Topic: record.Topic,
			}
		}

		success := false
		for attempt := 1; attempt <= 3; attempt++ {
			err := c.Processor.ProcessRecords(ctx, records)
			if err == nil {
				c.Logger.Info("successfully processed records", zap.Int("count", len(records)))
				success = true
				break
			}
			c.Logger.Warn("processing failed, retrying...", zap.Int("attempt", attempt), zap.Error(err))
			jitter := time.Duration(rand.Int63n(int64(time.Second)) * (1 << attempt)) // 1s, 2s-4s, 4s-8s, 8s-16s
			time.Sleep(jitter)
		}

		if !success {
			c.Logger.Info("processing failed after retries, sending to DLQ")
		}

		// Commit successfully processed records
		if err := c.Client.CommitRecords(ctx, fetches.Records()...); err != nil {
			c.Logger.Error("failed to commit processed records", zap.Error(err))
		}
	}
}
