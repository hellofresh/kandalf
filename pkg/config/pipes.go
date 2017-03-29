package config

import (
	"encoding/json"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Pipe contains settings for single bridge pipe between Kafka and RabbitMQ
type Pipe struct {
	KafkaTopic         string `yaml:"kafka_topic" json:"kafka_topic"`
	RabbitExchangeName string `yaml:"rabbitmq_exchange_name" json:"rabbitmq_exchange_name"`
	RabbitRoutingKey   string `yaml:"rabbitmq_routing_key" json:"rabbitmq_routing_key"`
	RabbitQueueName    string `yaml:"rabbitmq_queue_name" json:"rabbitmq_queue_name"`
}

func (p Pipe) String() string {
	b, _ := json.Marshal(p)
	return string(b)
}

// LoadPipesFromFile loads pipes config from file
func LoadPipesFromFile(pipesConfigPath string) ([]Pipe, error) {
	data, err := ioutil.ReadFile(pipesConfigPath)
	if err != nil {
		return nil, err
	}

	pipes, err := LoadPipesFromData(data)
	if err != nil {
		return nil, err
	}

	return pipes, nil
}

// LoadPipesFromData loads pipes config from data
func LoadPipesFromData(data []byte) ([]Pipe, error) {
	var pipes []Pipe

	if err := yaml.Unmarshal(data, &pipes); err != nil {
		return pipes, err
	}

	return pipes, nil
}
