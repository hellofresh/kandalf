package storage

import (
	"net/url"
	"time"

	"github.com/gomodule/redigo/redis"
)

// RedisStorage is a PersistentStorage interface implementation for Redis DB
type RedisStorage struct {
	pool *redis.Pool
	key  string
}

// NewRedisStorage instantiates and establishes connection to Redis storage
func NewRedisStorage(dsn *url.URL, key string) (*RedisStorage, error) {
	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.DialURL(dsn.String()) },
	}

	redisStorage := &RedisStorage{pool, key}

	conn := redisStorage.getConnection()
	defer conn.Close()
	if _, err := redisStorage.ping(conn); err != nil {
		return nil, err
	}

	return redisStorage, nil
}

func (s *RedisStorage) getConnection() redis.Conn {
	return s.pool.Get()
}

func (s *RedisStorage) ping(conn redis.Conn) (bool, error) {
	data, err := conn.Do("PING")
	if err != nil || data == nil {
		return false, err
	}

	return data == "PONG", nil
}

// Put writes data to Redis
func (s *RedisStorage) Put(data []byte) error {
	conn := s.getConnection()
	defer conn.Close()

	_, err := s.put(conn, data)
	return err
}

func (s *RedisStorage) put(conn redis.Conn, data []byte) (int, error) {
	return redis.Int(conn.Do("LPUSH", s.key, data))
}

// Get reads data from redis, if no more data in the storage "ErrStorageIsEmpty" is returned
func (s *RedisStorage) Get() ([]byte, error) {
	conn := s.getConnection()
	defer conn.Close()

	return s.get(conn)
}

func (s *RedisStorage) get(conn redis.Conn) ([]byte, error) {
	result, err := redis.Bytes(conn.Do("LPOP", s.key))
	if err == redis.ErrNil {
		return nil, ErrStorageIsEmpty
	}

	return result, err
}

// Close closes connection to redis
func (s *RedisStorage) Close() error {
	return s.pool.Close()
}
