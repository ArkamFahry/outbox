package internal

import (
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	ServiceEnvironmentDev  = "dev"
	ServiceEnvironmentTest = "test"
	ServiceEnvironmentProd = "prod"
)

type Config struct {
	ServiceName        string `json:"service_name" mapstructure:"service_name"`
	ServiceEnvironment string `json:"service_environment" mapstructure:"service_environment"`
	PostgresUrl        string `json:"postgresql_url" mapstructure:"postgresql_url"`
	NatsUrl            string `json:"nats_url" mapstructure:"nats_url"`
	OutboxTable        string `json:"outbox_table" mapstructure:"outbox_table"`
	PollingInterval    int    `json:"polling_interval" mapstructure:"polling_interval"`
	BatchSize          int    `json:"batch_size" mapstructure:"batch_size"`
}

func (c *Config) IsValid() error {
	if c.ServiceName == "" {
		return fmt.Errorf("service_name is required")
	}

	if c.ServiceEnvironment != "" {
		switch c.ServiceEnvironment {
		case ServiceEnvironmentDev, ServiceEnvironmentTest, ServiceEnvironmentProd:
		default:
			return fmt.Errorf("invalid service_environment '%s' service_environment can be dev, test or prod", c.ServiceEnvironment)
		}
	}

	if c.PostgresUrl == "" {
		return fmt.Errorf("postgresql_url is required")
	}

	if c.NatsUrl == "" {
		return fmt.Errorf("nats_url is required")
	}

	return nil
}

func (c *Config) SetDefaults() {
	if c.ServiceEnvironment == "" {
		c.ServiceEnvironment = ServiceEnvironmentProd
	}

	if c.OutboxTable == "" {
		c.OutboxTable = "events"
	}

	if c.PollingInterval == 0 {
		c.PollingInterval = 250
	}

	if c.BatchSize == 0 {
		c.BatchSize = 1000
	}
}

func NewConfig() *Config {
	var config Config

	v := viper.New()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	v.AddConfigPath(".")
	v.AddConfigPath("../")

	err = v.ReadInConfig()
	if err != nil {
		logger.Fatal("error reading config", zap.Error(err))
	}

	err = v.Unmarshal(&config)
	if err != nil {
		logger.Fatal("error unmarshaling config", zap.Error(err))
	}

	config.SetDefaults()

	err = config.IsValid()
	if err != nil {
		logger.Fatal("error validating config", zap.Error(err))
	}

	return &config
}
