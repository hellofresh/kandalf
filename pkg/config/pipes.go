package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Pipe struct {
	KafkaTopic         string `yaml:"kafka_topic"`
	RabbitExchangeName string `yaml:"rabbitmq_exchange_name"`
	RabbitRoutingKey   string `yaml:"rabbitmq_routing_key"`
	RabbitQueueName    string `yaml:"rabbitmq_queue_name"`
}

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
