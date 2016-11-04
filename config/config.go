package config

import (
	"log"
	"sync"

	l "github.com/hellofresh/kandalf/logger"
	c "github.com/olebedev/config"
)

var (
	config *c.Config
	mutex  *sync.Mutex = &sync.Mutex{}
)

// Returns the instance of config object.
// If the argument is provided, will try to reload config file
func Instance(paths ...string) *c.Config {
	if len(paths) > 0 {
		mutex.Lock()
		defer mutex.Unlock()

		config = getConfig(paths[0])
	}

	return config
}

// Instantiates the config object.
func getConfig(path string) *c.Config {
	var (
		err error
		cnf *c.Config
	)

	cnf, err = c.ParseYamlFile(path)
	if err != nil {
		log.Fatalf("Unable to read configuration file: %v", err)
	}

	// After reading config file, we need to setup logs again
	l.Close()
	l.Instance(cnf)

	return cnf
}
