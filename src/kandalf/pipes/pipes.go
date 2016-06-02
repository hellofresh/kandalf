package pipes

import (
	"io/ioutil"
	"log"
	"sort"
	"sync"

	"gopkg.in/yaml.v2"
)

type Pipe struct {
	Topic        string `yaml:"topic"`
	ExchangeName string `yaml:"exchange_name,omitempty"`
	RoutedQueue  string `yaml:"routed_queue,omitempty"`
	RoutingKey   string `yaml:"routing_key,omitempty"`
	Priority     int    `yaml:"priority,omitempty"`

	HasExchangeName bool
	HasRoutedQueue  bool
	HasRoutingKey   bool
}

type tPipes []Pipe

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
func getPipes(path string) (pipes tPipes) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Unable to read file with pipes: %v", err)
	}

	err = yaml.Unmarshal([]byte(data), &pipes)
	if err != nil {
		log.Fatalf("Unable to parse pipes: %v", err)
	}

	for _, pipe := range pipes {
		pipe.HasExchangeName = len(pipe.ExchangeName) > 0
		pipe.HasRoutedQueue = len(pipe.RoutedQueue) > 0
		pipe.HasRoutingKey = len(pipe.RoutingKey) > 0
	}

	sort.Sort(pipes)

	return pipes
}

// Method to satisfy sort.Interface
func (slice tPipes) Len() int {
	return len(slice)
}

// Method to satisfy sort.Interface
func (slice tPipes) Less(i, j int) bool {
	return slice[i].Priority > slice[j].Priority
}

// Method to satisfy sort.Interface
func (slice tPipes) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
