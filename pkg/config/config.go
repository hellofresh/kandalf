package config

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
)

type Duration struct {
	time.Duration
}

func (d Duration) Decode(value string) error {
	if val, err := time.ParseDuration(value); err != nil {
		return err
	} else {
		d.Duration = val
	}
	return nil
}

type GlobalConfig struct {
	LogLevel  string `envconfig:"LOG_LEVEL" default:"info"`
	RabbitDSN string `envconfig:"RABBIT_DSN" required:"true"`
	// StorageDSN is DSN for persistent storage used in case of Kafka unavailability. Example:
	//  redis://redis.local/?key=storage:key
	StorageDSN string `envconfig:"STORAGE_DSN" required:"true"`

	Kafka  KafkaConfig
	Stats  StatsConfig
	Worker WorkerConfig
}

type KafkaConfig struct {
	Brokers     []string `envconfig:"KAFKA_BROKERS" required:"true"`
	MaxRetry    int      `envconfig:"KAFKA_MAX_RETRY" default:"5"`
	PipesConfig string   `envconfig:"KAFKA_PIPES_CONFIG" default:"/etc/kandalf/pipes.yml"`
}

type StatsConfig struct {
	DSN    string `envconfig:"STATS_DSN"`
	Prefix string `envconfig:"STATS_PREFIX"`
}

type WorkerConfig struct {
	// CycleTimeout is worker cycle sleep time to avoid CPU overload
	CycleTimeout Duration `envconfig:"WORKER_CYCLE_TIMEOUT" default:"2s"`
	// CacheSize is max messages number that we store in memory before trying to publish to Kafka
	CacheSize int `envconfig:"WORKER_CACHE_SIZE" default:"10"`
	// CacheFlushTimeout is max amount of time we store messages in memory before trying to publish to Kafka
	CacheFlushTimeout Duration `envconfig:"WORKER_CACHE_FLUSH_TIMEOUT" default:"5s"`
	// ReadTimeout is timeout between attempts of reading persisted messages from storage
	// to publish them to Kafka, must be at least 2x greater than CycleTimeout
	StorageReadTimeout Duration `envconfig:"WORKER_STORAGE_READ_TIMEOUT" default:"10s"`
}

var instance GlobalConfig

func LoadEnv() GlobalConfig {
	err := envconfig.Process("", &instance)
	if err != nil {
		log.WithError(err).Fatal("Failed to load instance from environment")
	}

	return instance
}
