package config

import (
	"fmt"
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

	// non-supported file type
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

func TestPipe_String(t *testing.T) {
	pipe := Pipe{
		KafkaTopic:         "topic",
		RabbitExchangeName: "rqExchange",
		RabbitRoutingKey:   "rqKey",
		RabbitQueueName:    "rqQueue",
	}
	pipeJSON := `{"KafkaTopic":"topic","RabbitExchangeName":"rqExchange","RabbitRoutingKey":"rqKey","RabbitQueueName":"rqQueue"}`

	assert.Equal(t, pipeJSON, pipe.String())
	assert.Equal(t, pipeJSON, fmt.Sprintf("%s", pipe))
}
