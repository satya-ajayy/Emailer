package kafka

import (
	// Go Internal Packages
	"context"
	"errors"
	"fmt"

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
	client    *kgo.Client
	config    *ConsumerConfig
	processor MailProcessor
	logger    *zap.Logger
}

type MailProcessor interface {
	ProcessRecord(record models.Record) error
}

// NewConsumer creates a new consumer to consume kafka topic
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
		client:    client,
		config:    conf,
		processor: processor,
		logger:    logger,
	}, nil
}

func (c *Consumer) SendToDLQ(ctx context.Context, record models.Record) {
	c.logger.Info("processing failed, sending to DLQ")
	dlqRecord := &kgo.Record{
		Key:   record.Key,
		Value: record.Value,
		Topic: fmt.Sprintf("%s-dlq", record.Topic),
	}
	if err := c.client.ProduceSync(ctx, dlqRecord).FirstErr(); err != nil {
		c.logger.Error("failed to send DLQ record", zap.Error(err))
	}
}

// Poll polls for records from the Kafka broker.
func (c *Consumer) Poll(ctx context.Context) error {
	defer c.client.Close()

	for {
		// Check if the context is canceled before polling
		if ctx.Err() != nil {
			c.logger.Warn("polling stopped: context canceled")
			return ctx.Err() // Exit gracefully
		}

		c.logger.Info(fmt.Sprintf("%s: polling for records", c.config.Name))
		fetches := c.client.PollRecords(ctx, c.config.RecordsPerPoll)

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

		for _, record := range records {
			err := c.processor.ProcessRecord(record)
			if err != nil {
				c.SendToDLQ(ctx, record)
			}
		}
		c.logger.Info("processed records", zap.Int("records", len(records)))

		// Commit successfully processed records
		if err := c.client.CommitRecords(ctx, fetches.Records()...); err != nil {
			c.logger.Error("failed to commit processed records", zap.Error(err))
		}
	}
}
