package kafka

import (
	// Go Internal Packages
	"context"
	"errors"
	"fmt"

	// Local Packages
	config "emailer/config"
	models "emailer/models"
	slack "emailer/utils/slack"

	// External Packages
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kprom"
	"go.uber.org/zap"
)

type Consumer struct {
	client    *kgo.Client
	config    *models.ConsumerConfig
	processor MailProcessor
	logger    *zap.Logger
	slack     config.Slack
	isProd    bool
}

type MailProcessor interface {
	ProcessRecord(ctx context.Context, record models.Record) error
}

// NewConsumer creates a new consumer to consume mails
// (PS: Must call Poll to start consuming the records)
func NewConsumer(conf *models.ConsumerConfig, processor MailProcessor, metrics *kprom.Metrics, logger *zap.Logger, slack config.Slack, isProd bool) (*Consumer, error) {
	c := &Consumer{
		config:    conf,
		processor: processor,
		logger:    logger,
		slack:     slack,
		isProd:    isProd,
	}

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

	c.client = client
	return c, nil
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

		total := len(records)
		c.logger.Info("received records", zap.Int("records", total))

		success := total
		for _, record := range records {
			err := c.processor.ProcessRecord(ctx, record)
			if err != nil {
				c.logger.Error("failed to process record", zap.Error(err))
				sender := slack.NewSender(c.slack, c.isProd)
				if err = sender(record, err); err != nil {
					c.logger.Error("failed to send slack message", zap.Error(err))
				}
				success = success - 1
			}
		}

		// Commit successfully processed records
		c.logger.Info("processed records", zap.Int("success", success), zap.Int("failed", total-success))
		if err := c.client.CommitRecords(ctx, fetches.Records()...); err != nil {
			c.logger.Error("failed to commit processed records", zap.Error(err))
		}

		c.client.AllowRebalance()
	}
}

// Ping pings the kgo client
func (c *Consumer) Ping(ctx context.Context) error {
	return c.client.Ping(ctx)
}
