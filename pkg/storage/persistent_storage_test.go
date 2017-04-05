package storage

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestNewPersistentStorage_ErrUnknownStorage(t *testing.T) {
	dsn, _ := url.Parse("unknown://localhost")
	storage, err := NewPersistentStorage(dsn)
	assert.Nil(t, storage)
	assert.NotEmpty(t, err)
	assert.Equal(t, ErrUnknownStorage, err)
}

func TestNewPersistentStorage_ErrRedisKeyMissed(t *testing.T) {
	dsn, _ := url.Parse("redis://localhost/")
	storage, err := NewPersistentStorage(dsn)
	assert.Nil(t, storage)
	assert.NotEmpty(t, err)
	assert.Equal(t, ErrRedisKeyMissed, err)
}
