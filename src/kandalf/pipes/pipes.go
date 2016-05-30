package pipes

import (
	"io/ioutil"
	"log"
	"sync"

	"gopkg.in/yaml.v2"
)

type Pipe struct {
	Topic        string   `yaml:"topic"`
	Exchange     string   `yaml:"exchange,omitempty"`
	RoutingKeys  []string `yaml:"routing_keys,omitempty"`
	RoutedQueues []string `yaml:"routed_queues,omitempty"`
}

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
func getPipes(path string) (pipes []Pipe) {
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
