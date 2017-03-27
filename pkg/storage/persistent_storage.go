package storage

import (
	"errors"
	"net/url"

	log "github.com/Sirupsen/logrus"
)

var (
	ErrStorageIsEmpty = errors.New("No more messages in storage")
	ErrUnknownStorage = errors.New("Unknown storage type")
)

type PersistentStorage interface {
	Put(data []byte) error
	Get() ([]byte, error)
	Close() error
}

func NewPersistentStorage(dsn *url.URL) (PersistentStorage, error) {
	log.WithField("dsn", dsn.String()).Debug("Trying to instantiate new persistent storage instance")

	log.WithField("type", dsn.Scheme).Info("Looking for storage")
	switch dsn.Scheme {
	case "redis":
		if len(dsn.Query().Get("key")) < 1 {
			return nil, errors.New("Redis storage requires 'key' parameter")
		}
		return NewRedisStorage(dsn.Path, dsn.Query().Get("key"))
	}
	return nil, ErrUnknownStorage
}
