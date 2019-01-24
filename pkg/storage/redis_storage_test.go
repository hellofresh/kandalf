package storage

import (
	"errors"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/rafaeljusto/redigomock"
	"github.com/stretchr/testify/assert"
)

func TestRedisStorage_ping_ok(t *testing.T) {
	conn := redigomock.NewConn()
	cmd := conn.Command("PING").Expect("PONG")
	defer conn.Clear()

	redisStorage := &RedisStorage{}

	result, err := redisStorage.ping(conn)
	assert.Equal(t, 1, conn.Stats(cmd))
	assert.True(t, result)
	assert.Nil(t, err)
}

func TestRedisStorage_ping_nok(t *testing.T) {
	conn := redigomock.NewConn()
	cmd := conn.Command("PING").Expect("NOPE")
	defer conn.Clear()

	redisStorage := &RedisStorage{}

	result, err := redisStorage.ping(conn)
	assert.Equal(t, 1, conn.Stats(cmd))
	assert.False(t, result)
	assert.Nil(t, err)
}

func TestRedisStorage_ping_error(t *testing.T) {
	redisErr := errors.New("test redis error")

	conn := redigomock.NewConn()
	cmd := conn.Command("PING").ExpectError(redisErr)
	defer conn.Clear()

	redisStorage := &RedisStorage{}

	result, err := redisStorage.ping(conn)
	assert.Equal(t, 1, conn.Stats(cmd))
	assert.False(t, result)
	assert.NotEmpty(t, err)
	assert.Equal(t, redisErr, err)
}

func TestRedisStorage_put_ok(t *testing.T) {
	data := []byte("Some data")
	key := uuid.Must(uuid.NewV4()).String()

	conn := redigomock.NewConn()
	cmd := conn.Command("LPUSH", key, data).Expect(int64(1))
	defer conn.Clear()

	redisStorage := &RedisStorage{key: key}

	_, err := redisStorage.put(conn, data)
	assert.Equal(t, 1, conn.Stats(cmd))
	assert.Nil(t, err)
}

func TestRedisStorage_put_error(t *testing.T) {
	redisErr := errors.New("test redis error")
	data := []byte("Some data")
	key := uuid.Must(uuid.NewV4()).String()

	conn := redigomock.NewConn()
	cmd := conn.Command("LPUSH", key, data).ExpectError(redisErr)
	defer conn.Clear()

	redisStorage := &RedisStorage{key: key}

	_, err := redisStorage.put(conn, data)
	assert.Equal(t, 1, conn.Stats(cmd))
	assert.NotEmpty(t, err)
	assert.Equal(t, redisErr, err)
}

func TestRedisStorage_get_ok(t *testing.T) {
	data := []byte("Some data")

	conn := redigomock.NewConn()
	cmd := conn.Command("LPOP").Expect(data)
	defer conn.Clear()

	redisStorage := &RedisStorage{}

	result, err := redisStorage.get(conn)
	assert.Equal(t, 1, conn.Stats(cmd))
	assert.Nil(t, err)
	assert.Equal(t, data, result)
}

func TestRedisStorage_get_empty(t *testing.T) {
	key := uuid.Must(uuid.NewV4()).String()

	conn := redigomock.NewConn()
	cmd := conn.Command("LPOP", key).Expect(nil)
	defer conn.Clear()

	redisStorage := &RedisStorage{key: key}

	_, err := redisStorage.get(conn)
	assert.Equal(t, 1, conn.Stats(cmd))
	assert.NotEmpty(t, err)
	assert.Equal(t, ErrStorageIsEmpty, err)
}

func TestRedisStorage_get_error(t *testing.T) {
	redisErr := errors.New("test redis error")

	conn := redigomock.NewConn()
	cmd := conn.Command("LPOP").ExpectError(redisErr)
	defer conn.Clear()

	redisStorage := &RedisStorage{}

	_, err := redisStorage.get(conn)
	assert.Equal(t, 1, conn.Stats(cmd))
	assert.NotEmpty(t, err)
	assert.Equal(t, redisErr, err)
}
