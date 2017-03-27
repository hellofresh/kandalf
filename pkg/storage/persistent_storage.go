package storage

import (
	"errors"
	"net/url"

	log "github.com/Sirupsen/logrus"
)

var (
	// ErrStorageIsEmpty is an error raised in case when there are no more messages in the storage
	ErrStorageIsEmpty = errors.New("No more messages in storage")
	// ErrUnknownStorage is an error raised in case when trying to instantiate storage of unknown type
	ErrUnknownStorage = errors.New("Unknown storage type")
)

// PersistentStorage is an interface for persistent storage
type PersistentStorage interface {
	// Put writes data to persistent storage
	Put(data []byte) error
	// Get reads data from persistent storage, if no more data in the storage "ErrStorageIsEmpty" is returned
	Get() ([]byte, error)
	// Close closes connection to persistent storage
	Close() error
}

// NewPersistentStorage instantiates and establishes connection to persistent storage of given type
func NewPersistentStorage(dsn *url.URL) (PersistentStorage, error) {
	log.WithField("dsn", dsn.String()).Debug("Trying to instantiate new persistent storage instance")

	log.WithField("type", dsn.Scheme).Info("Looking for storage")
	switch dsn.Scheme {
	case "redis":
		if len(dsn.Query().Get("key")) < 1 {
			return nil, errors.New("Redis storage requires 'key' parameter")
		}
		return NewRedisStorage(dsn, dsn.Query().Get("key"))
	}
	return nil, ErrUnknownStorage
}
