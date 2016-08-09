package logger

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/olebedev/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Strip possible colors from logging entry
var formatter = &logrus.TextFormatter{DisableColors: true}

// Hook to handle writing to local log files.
type fileHook struct {
	levels   []logrus.Level
	lj       *lumberjack.Logger
	minLevel logrus.Level
}

// Creates a new hook to stores messages in plain text file
func newFileHook(c *config.Config, level logrus.Level) (*fileHook, error) {
	path, err := c.String("log.file.path")
	if err != nil {
		return nil, err
	}

	lj := &lumberjack.Logger{Filename: path}

	maxSize, err := c.Int("log.file.max_size")
	if err == nil {
		lj.MaxSize = maxSize
	}

	maxBackups, err := c.Int("log.file.max_backups")
	if err == nil {
		lj.MaxBackups = maxBackups
	}

	maxAge, err := c.Int("log.file.max_age")
	if err == nil {
		lj.MaxAge = maxAge
	}

	return &fileHook{
		levels:   []logrus.Level{level},
		lj:       lj,
		minLevel: level,
	}, nil
}

func (h *fileHook) Fire(entry *logrus.Entry) error {
	// We should write only messages with provided level
	if entry.Level < h.minLevel {
		return nil
	}

	// Store and the restore previous formatter
	oldFormatter := entry.Logger.Formatter
	entry.Logger.Formatter = formatter
	defer func() {
		entry.Logger.Formatter = oldFormatter
	}()

	msg, err := entry.String()

	if err != nil {
		return fmt.Errorf("Unable to generate string for log entry: %v", err)
	}

	h.lj.Write([]byte(msg))

	return nil
}

func (h *fileHook) Levels() []logrus.Level {
	return h.levels
}

func (h *fileHook) Close() error {
	fmt.Println("Close lumberjack")
	return h.lj.Close()
}
