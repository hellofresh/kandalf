package config

import (
	"io/ioutil"
	"time"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

// Duration is "time.Duration" wrapper for parsing it from string for environment variables
type Duration struct {
	time.Duration
}

// Decode decodes string value from environment variable to "time.Duration"
func (d *Duration) Decode(value string) error {
	val, err := time.ParseDuration(value)
	if err != nil {
		return err
	}
	d.Duration = val
	return nil
}

// UnmarshalYAML decodes string value from yaml config variable to "time.Duration"
func (d *Duration) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	var i interface{}
	if err := unmarshal(&i); err != nil {
		return err
	}

	if value, ok := i.(string); ok {
		if value == "" {
			return &yaml.TypeError{Errors: []string{"Empty string is invalid value for duration field"}}
		}
		val, err := time.ParseDuration(value)
		if err != nil {
			return &yaml.TypeError{Errors: []string{err.Error()}}
		}
		d.Duration = val

		return nil
	}

	return &yaml.TypeError{Errors: []string{"Failed to cast duration field to string"}}
}

// GlobalConfig contains application configuration values
type GlobalConfig struct {
	// LogLevel defines logging level for application, default is "info"
	LogLevel string `yaml:"log_level" envconfig:"LOG_LEVEL" default:"info"`
	// RabbitDSN is DSN for RabbitMQ instance to consume messages from
	RabbitDSN string `yaml:"rabbit_dsn" envconfig:"RABBIT_DSN" required:"true"`
	// StorageDSN is DSN for persistent storage used in case of Kafka unavailability. Example:
	//  redis://redis.local/?key=storage:key
	StorageDSN string `yaml:"storage_dsn" envconfig:"STORAGE_DSN" required:"true"`

	// Kafka contains configuration values for Kafka
	Kafka KafkaConfig
	// Stats contains configuration values for stats
	Stats StatsConfig
	// Worker contains configuration values for actual bridge worker
	Worker WorkerConfig
}

// KafkaConfig contains application configuration values for Kafka
type KafkaConfig struct {
	// Brokers is Kafka brokers comma-separated list, e.g. "192.168.0.1:9092,192.168.0.2:9092"
	Brokers []string `envconfig:"KAFKA_BROKERS" required:"true"`
	// MaxRetry is total number of times to retry sending a message to Kafka, default is 5
	MaxRetry int `yaml:"max_retry" envconfig:"KAFKA_MAX_RETRY" default:"5"`
	// PipesConfig is a path to rabbit-kafka bridge mappings config.
	// This must be YAML file with teh following structure:
	//
	//  ---
	//  - rabbitmq_exchange_name: "customers"     # Message from that RabbitMQ exchange
	//    rabbitmq_routing_key:   "order.created" # With that routing key
	//    kafka_topic:            "new-orders"    # Will be placed to that kafka topic
	//    # The queue name can be whatever you want, just keep it unique within pipes.
	//    # If you launch multiple kandalf instances they all will consume messages from that queue.
	//    rabbitmq_queue_name:    "kandalf-customers-order.created"
	//  - kafka_topic:            "loyalty"
	//    rabbitmq_exchange_name: "customers"
	//    rabbitmq_routing_key:   "badge.received"
	//    rabbitmq_queue_name:    "kandalf-customers-badge.received"
	//
	// Default path is "/etc/kandalf/conf/pipes.yml".
	PipesConfig string `yaml:"pipes_config" envconfig:"KAFKA_PIPES_CONFIG" default:"/etc/kandalf/conf/pipes.yml"`
}

// StatsConfig contains application configuration values for stats.
// For details - read docs for github.com/hellofresh/stats-go package
type StatsConfig struct {
	// DSN is stats service DSN
	DSN string `envconfig:"STATS_DSN"`
	// Prefix is stats prefix
	Prefix string `envconfig:"STATS_PREFIX"`
}

// WorkerConfig contains application configuration values for actual bridge worker
type WorkerConfig struct {
	// CycleTimeout is worker cycle sleep time to avoid CPU overload
	CycleTimeout Duration `yaml:"cycle_timeout" envconfig:"WORKER_CYCLE_TIMEOUT" default:"2s"`
	// CacheSize is max messages number that we store in memory before trying to publish to Kafka
	CacheSize int `yaml:"cache_size" envconfig:"WORKER_CACHE_SIZE" default:"10"`
	// CacheFlushTimeout is max amount of time we store messages in memory before trying to publish to Kafka
	CacheFlushTimeout Duration `yaml:"cache_flush_timeout" envconfig:"WORKER_CACHE_FLUSH_TIMEOUT" default:"5s"`
	// ReadTimeout is timeout between attempts of reading persisted messages from storage
	// to publish them to Kafka, must be at least 2x greater than CycleTimeout
	StorageReadTimeout Duration `yaml:"storage_read_timeout" envconfig:"WORKER_STORAGE_READ_TIMEOUT" default:"10s"`
}

var instance GlobalConfig

// LoadConfigFromEnv populates config values from environment variables
func LoadConfigFromEnv() (GlobalConfig, error) {
	err := envconfig.Process("", &instance)
	if err != nil {
		return instance, err
	}

	return instance, nil
}

// LoadConfigFromFile populates config values from file
func LoadConfigFromFile(configPath string) (GlobalConfig, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return instance, err
	}

	instance, err := LoadConfigFromData(data)
	if err != nil {
		return instance, err
	}

	return instance, nil
}

// LoadConfigFromData populates config values from data
func LoadConfigFromData(data []byte) (GlobalConfig, error) {
	if err := yaml.Unmarshal(data, &instance); err != nil {
		return instance, err
	}

	return instance, nil
}
