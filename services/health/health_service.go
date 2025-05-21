package health

import (
	// Go Internal Packages
	"context"

	// Local Packages
	kafka "emailer/kafka"

	// External Packages
	"go.uber.org/zap"
)

type HealthCheckService struct {
	logger        *zap.Logger
	kafkaConsumer *kafka.Consumer
}

// NewService creates a new HealthCheckService instance and returns the instance.
func NewService(logger *zap.Logger, consumer *kafka.Consumer) *HealthCheckService {
	return &HealthCheckService{
		logger:        logger,
		kafkaConsumer: consumer,
	}
}

// Health checks the health of the connections and returns true if all the connections are healthy.
func (h *HealthCheckService) Health(ctx context.Context) bool {
	// check consumer ping
	if consumerPingErr := h.kafkaConsumer.Ping(ctx); consumerPingErr != nil {
		h.logger.Error("kafka consumer ping failed", zap.Error(consumerPingErr))
		return false
	}

	return true
}
