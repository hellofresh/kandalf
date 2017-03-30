package storage

import (
	"net/url"

	"github.com/go-redis/redis"
)

// RedisStorage is a PersistentStorage interface implementation for Redis DB
type RedisStorage struct {
	rc  *redis.Client
	key string
}

// NewRedisStorage instantiates and establishes connection to Redis storage
func NewRedisStorage(dsn *url.URL, key string) (*RedisStorage, error) {
	options := &redis.Options{Addr: dsn.Host}
	if password, isSet := dsn.User.Password(); isSet {
		options.Password = password
	}

	rc := redis.NewClient(options)
	_, err := rc.Ping().Result()
	if err != nil {
		return nil, err
	}

	return &RedisStorage{rc, key}, nil
}

// Put writes data to Redis
func (s *RedisStorage) Put(data []byte) error {
	return s.rc.LPush(s.key, string(data)).Err()
}

// Get reads data from redis, if no more data in the storage "ErrStorageIsEmpty" is returned
func (s *RedisStorage) Get() ([]byte, error) {
	data, err := s.rc.LPop(s.key).Bytes()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, ErrStorageIsEmpty
	}

	return data, nil
}

// Close closes connection to redis
func (s *RedisStorage) Close() error {
	return s.rc.Close()
}
