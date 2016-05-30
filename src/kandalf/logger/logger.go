package logger

import (
	"log"
	"sync"

	"github.com/Sirupsen/logrus"
	ls "github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/olebedev/config"
)

var (
	logger *logrus.Logger
	mutex  *sync.Mutex = &sync.Mutex{}
)

// Returns the instance of logger object.
// If the argument is provided, will try to reload config of logger.
func Instance(configs ...*config.Config) *logrus.Logger {
	if len(configs) > 0 {
		mutex.Lock()
		defer mutex.Unlock()

		logger = getLogger(configs[0])
	}

	return logger
}

// Instantiates the logger object.
func getLogger(c *config.Config) *logrus.Logger {
	l := logrus.New()
	l.Level = getLevel(c)

	hook, _ := getLogstashHook(c)
	if hook != nil {
		l.Hooks.Add(hook)
	}

	return l
}

// Returns log level
func getLevel(c *config.Config) (lvl logrus.Level) {
	level, err := c.String("log.level")
	if err == nil {
		lvl, err = logrus.ParseLevel(level)
		if err != nil {
			lvl = logrus.InfoLevel
		}
	} else {
		lvl = logrus.InfoLevel
	}

	return lvl
}

// Returns hook to send all logs to logstash
func getLogstashHook(c *config.Config) (hook *ls.Hook, err error) {
	protocol, err := c.String("log.logstash.protocol")
	if err != nil {
		return nil, err
	}

	address, err := c.String("log.logstash.address")
	if err != nil {
		return nil, err
	}

	hook, err = ls.NewHook(protocol, address)
	if err != nil {
		log.Fatalf("Unable to instantiate logstash hook: %v", err)
	}

	return hook, nil
}
