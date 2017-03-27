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

// LoadPipes loads pipes config from file
func LoadPipes(pipesConfigPath string) ([]Pipe, error) {
	var pipes []Pipe

	data, err := ioutil.ReadFile(pipesConfigPath)
	if err != nil {
		return pipes, err
	}

	if err = yaml.Unmarshal([]byte(data), &pipes); err != nil {
		return pipes, err
	}

	return pipes, nil
}
