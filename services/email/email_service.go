package email

import (
	// Go Internal Packages
	"context"
	"encoding/json"
	"fmt"

	// Local Packages
	kafka "emailer/kafka"
	models "emailer/models"
)

type EmailService struct {
	consumer *kafka.Consumer
}

func NewService(consumer *kafka.Consumer) *EmailService {
	return &EmailService{consumer: consumer}
}

func (s *EmailService) AddToKafka(ctx context.Context, mailQP models.MailQP) error {
	key := []byte(mailQP.Order)
	value, err := json.Marshal(mailQP)
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %v", err)
	}

	err = s.consumer.ProduceRecord(ctx, key, value)
	if err != nil {
		return fmt.Errorf("failed to send record to kafka: %v", err)
	}
	return nil
}
