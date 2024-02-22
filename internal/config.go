package internal

import "fmt"

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
	RetryLimit         int    `json:"retry_limit" mapstructure:"retry_limit"`
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

	if c.RetryLimit == 0 {
		c.RetryLimit = 10
	}
}
