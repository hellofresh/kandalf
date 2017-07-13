package config

import (
	"time"

	"github.com/hellofresh/logging-go"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// GlobalConfig contains application configuration values
type GlobalConfig struct {
	// RabbitDSN is DSN for RabbitMQ instance to consume messages from
	RabbitDSN string `envconfig:"RABBIT_DSN"`
	// StorageDSN is DSN for persistent storage used in case of Kafka unavailability. Example:
	//  redis://redis.local/?key=storage:key
	StorageDSN string `envconfig:"STORAGE_DSN"`

	// Log contains configuration values logging
	Log logging.LogConfig
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
	Brokers []string `envconfig:"KAFKA_BROKERS"`
	// MaxRetry is total number of times to retry sending a message to Kafka, default is 5
	MaxRetry int `envconfig:"KAFKA_MAX_RETRY"`
	// PipesConfig is a path to rabbit-kafka bridge mappings config.
	// This must be YAML file with the following structure:
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
	PipesConfig string `envconfig:"KAFKA_PIPES_CONFIG"`
}

// StatsConfig contains application configuration values for stats.
// For details - read docs for github.com/hellofresh/stats-go package
type StatsConfig struct {
	DSN           string `envconfig:"STATS_DSN"`
	Prefix        string `envconfig:"STATS_PREFIX"`
	ErrorsSection string `envconfig:"STATS_ERRORS_SECTION"`
}

// WorkerConfig contains application configuration values for actual bridge worker
type WorkerConfig struct {
	// CycleTimeout is worker cycle sleep time to avoid CPU overload
	CycleTimeout time.Duration `envconfig:"WORKER_CYCLE_TIMEOUT"`
	// CacheSize is max messages number that we store in memory before trying to publish to Kafka
	CacheSize int `envconfig:"WORKER_CACHE_SIZE"`
	// CacheFlushTimeout is max amount of time we store messages in memory before trying to publish to Kafka
	CacheFlushTimeout time.Duration `envconfig:"WORKER_CACHE_FLUSH_TIMEOUT"`
	// ReadTimeout is timeout between attempts of reading persisted messages from storage
	// to publish them to Kafka, must be at least 2x greater than CycleTimeout
	StorageReadTimeout time.Duration `envconfig:"WORKER_STORAGE_READ_TIMEOUT"`
	// StorageMaxErrors is max storage read errors in a row before worker stops trying reading in current
	// read cycle. Next read cycle will be in "StorageReadTimeout" interval.
	StorageMaxErrors int `envconfig:"WORKER_STORAGE_MAX_ERRORS"`
}

func init() {
	viper.SetDefault("logLevel", "info")
	viper.SetDefault("kafka.maxRetry", 5)
	viper.SetDefault("kafka.pipesConfig", "/etc/kandalf/conf/pipes.yml")
	viper.SetDefault("worker.cycleTimeout", time.Second*time.Duration(2))
	viper.SetDefault("worker.cacheSize", 10)
	viper.SetDefault("worker.cacheFlushTimeout", time.Second*time.Duration(5))
	viper.SetDefault("worker.storageReadTimeout", time.Second*time.Duration(10))
	viper.SetDefault("worker.storageMaxErrors", 10)
	viper.SetDefault("stats.dsn", "log://")
	viper.SetDefault("stats.errorsSection", "error-log")

	logging.InitDefaults(viper.GetViper(), "log")
}

// Load loads config values from file,
// fallback to load from environment variables if file is not found or failed to read
func Load(configPath string) (*GlobalConfig, error) {
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("/etc/kandalf/conf")
		viper.AddConfigPath(".")
	}

	if err := viper.ReadInConfig(); err != nil {
		log.WithError(err).Warn("No config file found, loading config from environment variables")
		return LoadConfigFromEnv()
	}
	log.WithField("path", viper.ConfigFileUsed()).Info("Config loaded from file")

	var instance GlobalConfig
	if err := viper.Unmarshal(&instance); err != nil {
		return nil, err
	}

	return &instance, nil
}

// LoadConfigFromEnv loads config values from environment variables
func LoadConfigFromEnv() (*GlobalConfig, error) {
	var instance GlobalConfig

	if err := viper.Unmarshal(&instance); err != nil {
		return nil, err
	}

	err := envconfig.Process("", &instance)
	if err != nil {
		return &instance, err
	}

	return &instance, nil
}
