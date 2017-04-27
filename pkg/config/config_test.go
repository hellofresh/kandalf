package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertConfig(t *testing.T, globalConfig *GlobalConfig) {
	assert.Equal(t, "info", globalConfig.LogLevel)
	assert.Equal(t, "amqp://user:password@rmq", globalConfig.RabbitDSN)
	assert.Equal(t, "redis://redis.local/?key=storage:key", globalConfig.StorageDSN)

	assert.Len(t, globalConfig.Kafka.Brokers, 2)
	assert.Equal(t, "192.0.0.1:9092", globalConfig.Kafka.Brokers[0])
	assert.Equal(t, "192.0.0.2:9092", globalConfig.Kafka.Brokers[1])
	assert.Equal(t, 5, globalConfig.Kafka.MaxRetry)
	assert.Equal(t, "/etc/kandalf/conf/pipes.yml", globalConfig.Kafka.PipesConfig)

	assert.Equal(t, "statsd.local:8125", globalConfig.Stats.DSN)
	assert.Equal(t, "kandalf", globalConfig.Stats.Prefix)

	assert.Equal(t, "2s", globalConfig.Worker.CycleTimeout.String())
	assert.Equal(t, 10, globalConfig.Worker.CacheSize)
	assert.Equal(t, "5s", globalConfig.Worker.CacheFlushTimeout.String())
	assert.Equal(t, "10s", globalConfig.Worker.StorageReadTimeout.String())
	assert.Equal(t, 10, globalConfig.Worker.StorageMaxErrors)
}

func TestLoad(t *testing.T) {
	wd, err := os.Getwd()
	assert.Nil(t, err)
	assert.Contains(t, wd, "github.com/hellofresh/kandalf")

	// .../github.com/hellofresh/kandalf/pkg/config/../../assets/config.yml
	globalConfigPath := filepath.Join(wd, "..", "..", "assets", "config.yml")
	_, err = os.Stat(globalConfigPath)
	assert.Nil(t, err)

	globalConfig, err := Load(globalConfigPath)
	assert.Nil(t, err)

	assertConfig(t, globalConfig)
}

func setGlobalConfigEnv() {
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("RABBIT_DSN", "amqp://user:password@rmq")
	os.Setenv("STORAGE_DSN", "redis://redis.local/?key=storage:key")
	os.Setenv("KAFKA_BROKERS", "192.0.0.1:9092,192.0.0.2:9092")
	os.Setenv("KAFKA_MAX_RETRY", "5")
	os.Setenv("KAFKA_PIPES_CONFIG", "/etc/kandalf/conf/pipes.yml")
	os.Setenv("STATS_DSN", "statsd.local:8125")
	os.Setenv("STATS_PREFIX", "kandalf")
	os.Setenv("WORKER_CYCLE_TIMEOUT", "2s")
	os.Setenv("WORKER_CACHE_SIZE", "10")
	os.Setenv("WORKER_CACHE_FLUSH_TIMEOUT", "5s")
	os.Setenv("WORKER_STORAGE_READ_TIMEOUT", "10s")
	os.Setenv("WORKER_STORAGE_MAX_ERRORS", "10")
}

func TestLoad_fallbackToEnv(t *testing.T) {
	setGlobalConfigEnv()

	globalConfig, err := Load("")
	assert.Nil(t, err)

	assertConfig(t, globalConfig)
}

func TestLoadConfigFromEnv(t *testing.T) {
	setGlobalConfigEnv()

	globalConfig, err := LoadConfigFromEnv()
	assert.Nil(t, err)

	assertConfig(t, globalConfig)
}

func TestLoadConfigFromEnv_Duration(t *testing.T) {
	setGlobalConfigEnv()

	os.Setenv("WORKER_CYCLE_TIMEOUT", "22s")
	os.Setenv("WORKER_CACHE_FLUSH_TIMEOUT", "55s")
	os.Setenv("WORKER_STORAGE_READ_TIMEOUT", "100s")

	globalConfig, err := LoadConfigFromEnv()
	assert.Nil(t, err)

	assert.Equal(t, "22s", globalConfig.Worker.CycleTimeout.String())
	assert.Equal(t, "55s", globalConfig.Worker.CacheFlushTimeout.String())
	assert.Equal(t, "1m40s", globalConfig.Worker.StorageReadTimeout.String())

	os.Setenv("WORKER_CYCLE_TIMEOUT", "hello")

	globalConfig, err = LoadConfigFromEnv()
	assert.NotEmpty(t, err)
}
