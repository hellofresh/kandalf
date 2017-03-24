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
	LogLevel             string   `envconfig:"LOG_LEVEL" default:"info"`
	RabbitDSN            string   `envconfig:"RABBIT_DSN" required:"true"`
	InfiniteCycleTimeout Duration `envconfig:"CYCLE_TIMEOUT" default:"2s"`

	Kafka KafkaConfig
	Redis RedisConfig
	Stats StatsConfig
}

type KafkaConfig struct {
	Brokers     []string `envconfig:"KAFKA_BROKERS" required:"true"`
	MaxRetry    int      `envconfig:"KAFKA_MAX_RETRY" default:"5"`
	PipesConfig string   `envconfig:"KAFKA_PIPES_CONFIG" default:"/etc/kandalf/pipes.yml"`
}

type RedisConfig struct {
	Address string `envconfig:"REDIS_ADDRESS" required:"true"`
	KeyName string `envconfig:"REDIS_KEY_NAME" required:"true"`
	// PoolSize is max messages number that we store in memory before trying to publish to Kafka
	PoolSize int `envconfig:"REDIS_POOL_SIZE" default:"10"`
	// FlushTimeout max amount of time we store messages in memory before trying to publish to Kafka
	FlushTimeout Duration `envconfig:"REDIS_FLUSH_TIMEOUT" default:"5s"`
	// ReadTimeout is timeout between attempts of reading back-uped messages from redis
	// to publish them to Kafka
	ReadTimeout Duration `envconfig:"REDIS_READ_TIMEOUT" default:"10s"`
}

type StatsConfig struct {
	DSN    string `envconfig:"STATS_DSN"`
	Prefix string `envconfig:"STATS_PREFIX"`
}

var instance GlobalConfig

func LoadEnv() GlobalConfig {
	err := envconfig.Process("", &instance)
	if err != nil {
		log.WithError(err).Fatal("Failed to load instance from environment")
	}

	return instance
}
