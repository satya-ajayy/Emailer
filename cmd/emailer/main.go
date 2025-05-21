package main

import (
	// Go Internal Packages
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	// Local Packages
	config "emailer/config"
	shttp "emailer/http"
	kafka "emailer/kafka"
	models "emailer/models"
	health "emailer/services/health"
	processors "emailer/services/processors"
	slack "emailer/utils/slack"

	// External Packages
	"github.com/alecthomas/kingpin/v2"
	_ "github.com/jsternberg/zap-logfmt"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/twmb/franz-go/plugin/kprom"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitializeServer sets up an HTTP server with health Service and starts kafka consumer.
func InitializeServer(ctx context.Context, k config.Config, logger *zap.Logger) (*shttp.Server, error) {
	// Slack Alert Sender
	slackAlerter := slack.NewSender(k.Slack, k.IsProdMode)

	metrics := kprom.NewMetrics("emailer")
	conf := &models.ConsumerConfig{
		Brokers:        k.Kafka.Brokers,
		Name:           k.Kafka.ConsumerName,
		Topic:          k.Kafka.Topic,
		RecordsPerPoll: k.Kafka.RecordsPerPoll,
	}

	processor := processors.NewProcessor(logger, k.Credentials)
	consumer, err := kafka.NewConsumer(conf, processor, metrics, logger, slackAlerter)
	if err != nil {
		return nil, err
	}

	go func() {
		if err = consumer.Poll(ctx); err != nil {
			logger.Fatal("cannot poll records from topic", zap.Error(err))
		}
	}()

	healthSvc := health.NewService(logger, consumer)
	server := shttp.NewServer(k.Prefix, logger, consumer, healthSvc)
	return server, nil
}

// LoadConfig loads the default configuration and overrides it with the config file
// specified by the path defined in the config flag
func LoadConfig() *koanf.Koanf {
	configPathMsg := "path to the application config file"
	configPath := kingpin.Flag("config", configPathMsg).Short('c').Default("config.yml").String()

	kingpin.Parse()
	k := koanf.New(".")
	_ = k.Load(rawbytes.Provider(config.DefaultConfig), yaml.Parser())
	if *configPath != "" {
		_ = k.Load(file.Provider(*configPath), yaml.Parser())
	}

	return k
}

func main() {
	k := LoadConfig()
	appKonf := config.Config{}

	// Unmarshalling config into struct
	err := k.Unmarshal("", &appKonf)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	// Validate the config loaded
	if err = appKonf.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	if !appKonf.IsProdMode {
		k.Print()
	}

	if !appKonf.Kafka.Consume {
		log.Fatalf("kafka consumer is not enabled")
	}

	cfg := zap.NewProductionConfig()
	cfg.Encoding = "logfmt"
	_ = cfg.Level.UnmarshalText([]byte(appKonf.Logger.Level))
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.InitialFields = make(map[string]any)
	cfg.InitialFields["host"], _ = os.Hostname()
	cfg.InitialFields["service"] = appKonf.Application
	cfg.OutputPaths = []string{"stdout"}
	logger, _ := cfg.Build()
	defer func() {
		_ = logger.Sync()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv, err := InitializeServer(ctx, appKonf, logger)
	if err != nil {
		logger.Fatal("cannot initialize server", zap.Error(err))
	}

	if err = srv.Listen(ctx, appKonf.Listen); err != nil {
		logger.Fatal("cannot listen", zap.Error(err))
	}
}
