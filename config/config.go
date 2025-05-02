package config

import (
	// Local Packages
	errors "emailer/errors"
)

var DefaultConfig = []byte(`
application: "emailer"

logger:
  level: "debug"

listen: ":2529"

prefix: "/emailer"

is_prod_mode: false

mongo:
  uri: "mongodb://localhost:27017"

kafka:
  brokers:
    - "localhost:9092"
  consume: true
  topic: "emails-to-send"
  records_per_poll: 50
  consumer_name: "emailer"

slack:
  webhook_url: "https://hooks.slack.com/services/your/webhook/url"
  send_alert_in_dev: true

credentials:
  mail_id: "your-email@example.com"
  password: "super-secret-password"
`)

type Config struct {
	Application string      `koanf:"application"`
	Listen      string      `koanf:"listen"`
	Prefix      string      `koanf:"prefix"`
	Logger      Logger      `koanf:"logger"`
	IsProdMode  bool        `koanf:"is_prod_mode"`
	Mongo       Mongo       `koanf:"mongo"`
	Slack       Slack       `koanf:"slack"`
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

type Mongo struct {
	URI string `koanf:"uri"`
}

type Slack struct {
	WebhookURL     string `koanf:"webhook_url"`
	SendAlertInDev bool   `koanf:"send_alert_in_dev"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
	ve := errors.ValidationErrs()

	if c.Application == "" {
		ve.Add("application", "cannot be empty")
	}
	if c.Listen == "" {
		ve.Add("listen", "cannot be empty")
	}
	if c.Logger.Level == "" {
		ve.Add("logger.level", "cannot be empty")
	}
	if c.Prefix == "" {
		ve.Add("prefix", "cannot be empty")
	}
	if c.Mongo.URI == "" {
		ve.Add("mongo.uri", "cannot be empty")
	}
	if c.Slack.WebhookURL == "" {
		ve.Add("slack.webhook_url", "cannot be empty")
	}

	return ve.Err()
}
