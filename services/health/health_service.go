package health

import (
	// Go Internal Packages
	"context"

	// Local Packages
	kafka "emailer/kafka"

	// External Packages
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type HealthCheckService struct {
	logger        *zap.Logger
	mongoClient   *mongo.Client
	kafkaConsumer *kafka.Consumer
}

// NewService creates a new HealthCheckService instance and returns the instance.
func NewService(logger *zap.Logger, mongoClient *mongo.Client, consumer *kafka.Consumer) *HealthCheckService {
	return &HealthCheckService{
		logger:        logger,
		mongoClient:   mongoClient,
		kafkaConsumer: consumer,
	}
}

// Health checks the health of the connections and returns true if all the connections are healthy.
func (h *HealthCheckService) Health(ctx context.Context) bool {
	// check mongo ping
	if mongoPingErr := h.mongoClient.Ping(ctx, nil); mongoPingErr != nil {
		h.logger.Error("mongo ping failed", zap.Error(mongoPingErr))
		return false
	}

	// check consumer ping
	if consumerPingErr := h.kafkaConsumer.Ping(ctx); consumerPingErr != nil {
		h.logger.Error("kafka consumer ping failed", zap.Error(consumerPingErr))
		return false
	}

	return true
}
