package logger

import (
	"log"
	"log/syslog"
	"reflect"
	"sync"

	"github.com/Sirupsen/logrus"
	hSyslog "github.com/Sirupsen/logrus/hooks/syslog"
	hLogstash "github.com/bshuster-repo/logrus-logstash-hook"
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

// Close all closable hooks on SIGHUP
func Close() {
	var closableType reflect.Type = reflect.TypeOf((*closable)(nil)).Elem()

	if logger != nil {
		for _, hooks := range logger.Hooks {
			for _, hook := range hooks {
				if reflect.TypeOf(hook).Implements(closableType) {
					_ = hook.(closable).Close()
				}
			}
		}
	}
}

// Instantiates the logger object.
func getLogger(c *config.Config) *logrus.Logger {
	var hook logrus.Hook

	l := logrus.New()
	l.Level = getLevel(c)

	adapter := c.UString("log.adapter", "syslog")
	switch adapter {
	case "file":
		hook, _ = getFileHook(c, l.Level)
	case "logstash":
		hook, _ = getLogstashHook(c)
	case "syslog":
	default:
		hook, _ = getSyslogHook(c, l.Level)
	}

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

// Returns hook that stores messages in plain text file
func getFileHook(c *config.Config, l logrus.Level) (hook *fileHook, err error) {
	hook, err = newFileHook(c, l)
	if err != nil {
		log.Fatalf("Unable to instantiate file logger hook: %v", err)
	}

	return hook, nil
}

// Returns hook to send all logs to logstash
func getLogstashHook(c *config.Config) (hook *hLogstash.Hook, err error) {
	protocol, err := c.String("log.logstash.protocol")
	if err != nil {
		return nil, err
	}

	address, err := c.String("log.logstash.address")
	if err != nil {
		return nil, err
	}

	hook, err = hLogstash.NewHook(protocol, address, "kandalf")
	if err != nil {
		log.Fatalf("Unable to instantiate logstash hook: %v", err)
	}

	return hook, nil
}

// Returns hook that sends all logs to syslog
func getSyslogHook(c *config.Config, l logrus.Level) (hook *hSyslog.SyslogHook, err error) {
	var priority syslog.Priority

	switch l {
	case logrus.DebugLevel:
		priority = syslog.LOG_DEBUG
	case logrus.InfoLevel:
		priority = syslog.LOG_INFO
	case logrus.WarnLevel:
		priority = syslog.LOG_WARNING
	case logrus.ErrorLevel:
		priority = syslog.LOG_ERR
	case logrus.FatalLevel:
		priority = syslog.LOG_CRIT
	case logrus.PanicLevel:
		priority = syslog.LOG_EMERG
	}

	protocol, err := c.String("log.syslog.protocol")
	if err != nil {
		return nil, err
	}

	address, err := c.String("log.syslog.address")
	if err != nil {
		return nil, err
	}

	hook, err = hSyslog.NewSyslogHook(protocol, address, priority, "kandalf")
	if err != nil {
		log.Fatalf("Unable to instantiate syslog hook: %v", err)
	}

	return hook, nil
}
