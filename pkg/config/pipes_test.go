package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertPipes(t *testing.T, pipes []Pipe) {
	assert.Equal(t, "customers", pipes[0].RabbitExchangeName)
	assert.Equal(t, []string{"order.created"}, pipes[0].RabbitRoutingKey)
	assert.Equal(t, "new-orders", pipes[0].KafkaTopic)
	assert.Equal(t, "kandalf-customers-order.created", pipes[0].RabbitQueueName)
	assert.Equal(t, true, pipes[0].RabbitDurableQueue)
	assert.Equal(t, false, pipes[0].RabbitAutoDeleteQueue)

	assert.Equal(t, "customers", pipes[1].RabbitExchangeName)
	assert.Equal(t, []string{"badge.received"}, pipes[1].RabbitRoutingKey)
	assert.Equal(t, "loyalty", pipes[1].KafkaTopic)
	assert.Equal(t, "kandalf-customers-badge.received", pipes[1].RabbitQueueName)
	assert.Equal(t, false, pipes[1].RabbitDurableQueue)
	assert.Equal(t, true, pipes[1].RabbitAutoDeleteQueue)

	assert.Equal(t, "users", pipes[2].RabbitExchangeName)
	assert.Equal(t, []string{"user.de.registered", "user.at.registered", "user.ch.registered"}, pipes[2].RabbitRoutingKey)
	assert.Equal(t, "topic_for_several_events", pipes[2].KafkaTopic)
	assert.Equal(t, "kandalf-users.user.registered", pipes[2].RabbitQueueName)
	assert.Equal(t, true, pipes[2].RabbitDurableQueue)
	assert.Equal(t, false, pipes[2].RabbitAutoDeleteQueue)
}

func TestLoadPipesFromFile(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	assert.Contains(t, wd, "github.com/hellofresh/kandalf")

	// .../github.com/hellofresh/kandalf/pkg/config/../../assets/pipes.yml
	pipesPath := filepath.Join(wd, "..", "..", "assets", "pipes.yml")
	_, err = os.Stat(pipesPath)
	assert.Nil(t, err)

	pipes, err := LoadPipesFromFile(pipesPath)
	require.NoError(t, err)
	assert.Len(t, pipes, 3)

	assertPipes(t, pipes)
}

func TestLoadPipesFromFile_Error(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	assert.Contains(t, wd, "github.com/hellofresh/kandalf")

	// non-supported file type
	// .../github.com/hellofresh/kandalf/pkg/config/pipes.go
	pipesPath := filepath.Join(wd, "pipes.go")
	_, err = os.Stat(pipesPath)
	require.NoError(t, err)

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

func TestPipe_String(t *testing.T) {
	pipe := Pipe{
		KafkaTopic:            "topic",
		RabbitExchangeName:    "rqExchange",
		RabbitRoutingKey:      []string{"rqKey"},
		RabbitQueueName:       "rqQueue",
		RabbitDurableQueue:    true,
		RabbitAutoDeleteQueue: false,
	}
	pipeJSON := `{"KafkaTopic":"topic","RabbitExchangeName":"rqExchange","RabbitRoutingKey":["rqKey"],"RabbitQueueName":"rqQueue","RabbitDurableQueue":true,"RabbitAutoDeleteQueue":false}`

	assert.Equal(t, pipeJSON, pipe.String())
	assert.Equal(t, pipeJSON, fmt.Sprintf("%s", pipe))
}
