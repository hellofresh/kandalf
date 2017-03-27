package storage

import "gopkg.in/redis.v3"

type RedisStorage struct {
	rc  *redis.Client
	key string
}

func NewRedisStorage(redisAddress string, key string) (*RedisStorage, error) {
	rc := redis.NewClient(&redis.Options{Addr: redisAddress})
	_, err := rc.Ping().Result()
	if err != nil {
		return nil, err
	}

	return &RedisStorage{rc, key}, nil
}

func (s *RedisStorage) Put(data []byte) error {
	return s.rc.LPush(s.key, string(data)).Err()
}

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

func (s *RedisStorage) Close() error {
	return s.rc.Close()
}
