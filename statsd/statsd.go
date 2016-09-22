package statsd

import (
	"sync"

	"github.com/olebedev/config"
	s "gopkg.in/alexcesaro/statsd.v2"

	log "kandalf/logger"
)

var (
	client *s.Client
	err    error
	mutex  *sync.Mutex = &sync.Mutex{}
)

// Returns the instance of StatsD client.
// If the argument is provided, will try to reload client.
func Instance(configs ...*config.Config) *s.Client {
	if len(configs) > 0 {
		mutex.Lock()
		defer mutex.Unlock()

		if client != nil {
			client.Close()
		}

		client, err = getClient(configs[0])
		if err != nil {
			log.Instance().
				WithError(err).
				Fatal("Unable to instantiate StatsD client")
		}
	}

	return client
}

// Instantiates the client object.
func getClient(c *config.Config) (*s.Client, error) {
	var options []s.Option

	address := c.UString("statsd.address", "")
	prefix := c.UString("statsd.prefix", "")

	if len(address) == 0 {
		options = append(options, s.Mute(true))
	} else {
		options = append(options, s.Address(address))
	}

	if len(prefix) > 0 {
		options = append(options, s.Prefix(prefix))
	}

	return s.New(options...)
}
