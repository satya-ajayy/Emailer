package config

import (
	// Local Packages
	errors "emailer/errors"
)

var DefaultConfig = []byte(`
application: "emailer"

logger:
  level: "debug"

is_prod_mode: false

kafka:
  brokers:
    - "localhost:9092"
  consume: true
  topic: "emails-to-send"
  records_per_poll: 50
  consumer_name: "emailer"

credentials:
  mail_id: "your-email@example.com"
  password: "super-secret-password"
`)

type Config struct {
	Application string      `koanf:"application"`
	Logger      Logger      `koanf:"logger"`
	IsProdMode  bool        `koanf:"is_prod_mode"`
	Kafka       Kafka       `koanf:"kafka"`
	Credentials Credentials `koanf:"credentials"`
}

type Logger struct {
	Level string `koanf:"level"`
}

type Kafka struct {
	Brokers        []string `koanf:"brokers"`
	Consume        bool     `koanf:"consume"`
	Topic          string   `koanf:"topic"`
	RecordsPerPoll int      `koanf:"records_per_poll"`
	ConsumerName   string   `koanf:"consumer_name"`
}

type Credentials struct {
	MailID   string `koanf:"mail_id"`
	Password string `koanf:"password"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
	ve := errors.ValidationErrs()

	if c.Application == "" {
		ve.Add("application", "cannot be empty")
	}
	if c.Logger.Level == "" {
		ve.Add("logger.level", "cannot be empty")
	}
	if len(c.Kafka.Brokers) == 0 {
		ve.Add("kafka.brokers", "cannot be empty")
	}

	return ve.Err()
}
