package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertPipes(t *testing.T, pipes []Pipe) {
	assert.Equal(t, "customers", pipes[0].RabbitExchangeName)
	assert.Equal(t, "order.created", pipes[0].RabbitRoutingKey)
	assert.Equal(t, "new-orders", pipes[0].KafkaTopic)
	assert.Equal(t, "kandalf-customers-order.created", pipes[0].RabbitQueueName)

	assert.Equal(t, "customers", pipes[1].RabbitExchangeName)
	assert.Equal(t, "badge.received", pipes[1].RabbitRoutingKey)
	assert.Equal(t, "loyalty", pipes[1].KafkaTopic)
	assert.Equal(t, "kandalf-customers-badge.received", pipes[1].RabbitQueueName)
}

func TestLoadPipesFromData(t *testing.T) {
	data := []byte(`
---
-
  # Message from that RabbitMQ exchange
  rabbitmq_exchange_name: "customers"
  # With that routing key
  rabbitmq_routing_key: "order.created"
  # Will be placed to that kafka topic
  kafka_topic: "new-orders"
  # The queue name can be whatever you want, just keep it unique within pipes.
  # If you launch multiple kandalf instances they all will consume messages from that queue.
  rabbitmq_queue_name: "kandalf-customers-order.created"

-
  kafka_topic: "loyalty"
  rabbitmq_exchange_name: "customers"
  rabbitmq_routing_key: "badge.received"
  rabbitmq_queue_name: "kandalf-customers-badge.received"
  `)
	pipes, err := LoadPipesFromData(data)
	assert.Nil(t, err)
	assert.Len(t, pipes, 2)

	assertPipes(t, pipes)
}

func TestLoadPipesFromData_Error(t *testing.T) {
	data := []byte("this is not yaml")
	_, err := LoadPipesFromData(data)
	assert.NotEmpty(t, err)
}

func TestLoadPipesFromFile(t *testing.T) {
	wd, err := os.Getwd()
	assert.Nil(t, err)
	assert.Contains(t, wd, "github.com/hellofresh/kandalf")

	// .../github.com/hellofresh/kandalf/pkg/config/../../ci/assets/pipes.yml
	pipesPath := filepath.Join(wd, "..", "..", "ci", "assets", "pipes.yml")
	_, err = os.Stat(pipesPath)
	assert.Nil(t, err)

	pipes, err := LoadPipesFromFile(pipesPath)
	assert.Nil(t, err)
	assert.Len(t, pipes, 2)

	assertPipes(t, pipes)
}

func TestLoadPipesFromFile_Error(t *testing.T) {
	wd, err := os.Getwd()
	assert.Nil(t, err)
	assert.Contains(t, wd, "github.com/hellofresh/kandalf")

	// non-yaml file
	// .../github.com/hellofresh/kandalf/pkg/config/pipes.go
	pipesPath := filepath.Join(wd, "pipes.go")
	_, err = os.Stat(pipesPath)
	assert.Nil(t, err)

	_, err = LoadPipesFromFile(pipesPath)
	assert.NotEmpty(t, err)

	// non-existent file
	// .../github.com/hellofresh/kandalf/pkg/config/does-not-exist.yml
	pipesPath = filepath.Join(wd, "does-not-exist.yml")
	_, err = os.Stat(pipesPath)
	assert.True(t, os.IsNotExist(err))

	_, err = LoadPipesFromFile(pipesPath)
	assert.NotEmpty(t, err)
}
