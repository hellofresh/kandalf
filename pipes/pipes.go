package pipes

import (
	"io/ioutil"
	"log"
	"sync"

	"gopkg.in/yaml.v2"
)

type Pipe struct {
	KafkaTopic           string `yaml:"kafka_topic"`
	RabbitmqExchangeName string `yaml:"rabbitmq_exchange_name"`
	RabbitmqRoutingKey   string `yaml:"rabbitmq_routing_key"`
	RabbitmqQueueName    string `yaml:"rabbitmq_queue_name"`
}

type PipesList []Pipe

var (
	pipes []Pipe
	mutex *sync.Mutex = &sync.Mutex{}
)

// Returns list of available pipes
func All(paths ...string) []Pipe {
	if len(paths) > 0 {
		mutex.Lock()
		defer mutex.Unlock()

		pipes = getPipes(paths[0])
	}

	return pipes
}

// Reads file with pipes in YML and returns list of pipes
func getPipes(path string) (pipes PipesList) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Unable to read file with pipes: %v", err)
	}

	err = yaml.Unmarshal([]byte(data), &pipes)
	if err != nil {
		log.Fatalf("Unable to parse pipes: %v", err)
	}

	return pipes
}
