package config

import (
	"encoding/json"
	"github.com/spf13/viper"
)

// Pipe contains settings for single bridge pipe between Kafka and RabbitMQ
type Pipe struct {
	KafkaTopic         string
	RabbitExchangeName string
	RabbitRoutingKey   string
	RabbitQueueName    string
}

func (p Pipe) String() string {
	b, _ := json.Marshal(p)
	return string(b)
}

// LoadPipesFromFile loads pipes config from file
func LoadPipesFromFile(pipesConfigPath string) ([]Pipe, error) {
	pipesConfigReader := viper.New()
	pipesConfigReader.SetConfigFile(pipesConfigPath)
	if err := pipesConfigReader.ReadInConfig(); err != nil {
		return nil, err
	}

	var pipes struct {
		Pipes []Pipe
	}

	if err := pipesConfigReader.Unmarshal(&pipes); err != nil {
		return nil, err
	}

	return pipes.Pipes, nil
}
