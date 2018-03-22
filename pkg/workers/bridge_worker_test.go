package workers

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/hellofresh/kandalf/pkg/config"
	"github.com/hellofresh/kandalf/pkg/producer"
	"github.com/hellofresh/kandalf/pkg/storage"
	"github.com/hellofresh/stats-go"
	"github.com/hellofresh/stats-go/client"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

type mockGetResult struct {
	data []byte
	err  error
}

type mockStorage struct {
	t *testing.T

	getResult   []mockGetResult
	putResult   []error
	closeResult error

	putData [][]byte

	getCalled int
	putCalled int
}

func (s *mockStorage) Put(data []byte) error {
	methodCall := s.putCalled
	s.putData = append(s.putData, data)

	s.putCalled++
	return s.putResult[methodCall]
}

func (s *mockStorage) Get() ([]byte, error) {
	methodCall := s.getCalled
	if methodCall+1 > len(s.getResult) {
		return nil, storage.ErrStorageIsEmpty
	}
	s.getCalled++

	return s.getResult[methodCall].data, s.getResult[methodCall].err
}

func (s *mockStorage) Close() error {
	return s.closeResult
}

type mockProducer struct {
	t *testing.T

	publishAssertParam []producer.Message
	publishResult      []error

	publishCalled int
}

func (p *mockProducer) Publish(msg producer.Message) error {
	methodCall := p.publishCalled
	assert.False(p.t, methodCall+1 > len(p.publishAssertParam))
	assert.Equal(p.t, p.publishAssertParam[methodCall], msg)

	p.publishCalled++

	return p.publishResult[methodCall]
}

func (p *mockProducer) Close() error {
	return nil
}

func generateRandomMessages(n int) []*producer.Message {
	result := make([]*producer.Message, n)
	for i := 0; i < n; i++ {
		result[i] = producer.NewMessage([]byte(uuid.NewV4().String()), uuid.NewV4().String())
	}
	return result
}

func TestBridgeWorker_MessageHandler(t *testing.T) {
	workerConfig := config.WorkerConfig{}
	statsClient, _ := stats.NewClient("memory://")
	mockStorage := &mockStorage{}
	mockProducer := &mockProducer{}

	messagesToPublish := 5

	messages := generateRandomMessages(messagesToPublish)
	worker, _ := NewBridgeWorker(workerConfig, mockStorage, mockProducer, statsClient)
	for _, msg := range messages {
		worker.MessageHandler(msg.Body, config.Pipe{KafkaTopic: msg.Topic})
	}

	memoryStats, _ := statsClient.(*client.Memory)
	assert.Equal(t, messagesToPublish, memoryStats.CountMetrics[fmt.Sprintf("total.%s", statsWorkerSection)])
	assert.Equal(t, messagesToPublish, memoryStats.CountMetrics[fmt.Sprintf("total.%s-ok", statsWorkerSection)])
	assert.Equal(t, 0, memoryStats.CountMetrics[fmt.Sprintf("total.%s-fail", statsWorkerSection)])

	assert.Equal(t, messagesToPublish, len(worker.cache))
	for i, msg := range messages {
		assert.Equal(t, msg.Topic, worker.cache[i].Topic)
		assert.Equal(t, msg.Body, worker.cache[i].Body)
		assert.NotEqual(t, msg.ID.String(), worker.cache[i].ID.String())

		assert.Equal(t, 1, memoryStats.CountMetrics[fmt.Sprintf("%s.cache.add.%s", statsWorkerSection, msg.Topic)])
		assert.Equal(t, 1, memoryStats.CountMetrics[fmt.Sprintf("%s-ok.cache.add.%s", statsWorkerSection, msg.Topic)])
		assert.Equal(t, 0, memoryStats.CountMetrics[fmt.Sprintf("%s-fail.cache.add.%s", statsWorkerSection, msg.Topic)])
	}
}

func TestBridgeWorker_cacheMessage(t *testing.T) {
	workerConfig := config.WorkerConfig{}
	statsClient, _ := stats.NewClient("memory://")
	mockStorage := &mockStorage{}
	mockProducer := &mockProducer{}

	messagesToPublish := 5

	messages := generateRandomMessages(messagesToPublish)
	worker, _ := NewBridgeWorker(workerConfig, mockStorage, mockProducer, statsClient)
	for _, msg := range messages {
		worker.cacheMessage(msg)
	}

	assert.Equal(t, messagesToPublish, len(worker.cache))
	assert.Equal(t, messages, worker.cache)

	memoryStats, _ := statsClient.(*client.Memory)
	assert.Equal(t, messagesToPublish, memoryStats.CountMetrics[fmt.Sprintf("total.%s", statsWorkerSection)])
	assert.Equal(t, messagesToPublish, memoryStats.CountMetrics[fmt.Sprintf("total.%s-ok", statsWorkerSection)])
	assert.Equal(t, 0, memoryStats.CountMetrics[fmt.Sprintf("total.%s-fail", statsWorkerSection)])

	for _, msg := range messages {
		assert.Equal(t, 1, memoryStats.CountMetrics[fmt.Sprintf("%s.cache.add.%s", statsWorkerSection, msg.Topic)])
		assert.Equal(t, 1, memoryStats.CountMetrics[fmt.Sprintf("%s-ok.cache.add.%s", statsWorkerSection, msg.Topic)])
		assert.Equal(t, 0, memoryStats.CountMetrics[fmt.Sprintf("%s-fail.cache.add.%s", statsWorkerSection, msg.Topic)])
	}
}

func getDefaultBridgeWorker(t *testing.T) *BridgeWorker {
	workerConfig := config.WorkerConfig{
		CacheSize:          5,
		StorageMaxErrors:   10,
		CycleTimeout:       time.Duration(1) * time.Hour,
		CacheFlushTimeout:  time.Duration(1) * time.Hour,
		StorageReadTimeout: time.Duration(1) * time.Hour,
	}
	statsClient, _ := stats.NewClient("memory://")
	mockStorage := &mockStorage{t: t}
	mockProducer := &mockProducer{t: t}

	worker, _ := NewBridgeWorker(workerConfig, mockStorage, mockProducer, statsClient)
	return worker
}

func TestBridgeWorker_Execute_emptyCache(t *testing.T) {
	worker := getDefaultBridgeWorker(t)

	mockProducer := &mockProducer{t: t}
	worker.producer = mockProducer

	worker.Execute()
}

func TestBridgeWorker_Execute_noFlushTimeNoCacheSize(t *testing.T) {
	worker := getDefaultBridgeWorker(t)

	mockProducer := &mockProducer{t: t}
	worker.producer = mockProducer

	worker.lastFlush = time.Now()
	worker.cache = generateRandomMessages(worker.config.CacheSize - 1)
	worker.Execute()
}

func TestBridgeWorker_Execute_noFlushTimeCacheSize(t *testing.T) {
	worker := getDefaultBridgeWorker(t)

	mockProducer := &mockProducer{t: t}
	worker.producer = mockProducer

	worker.lastFlush = time.Now()
	messagesCount := worker.config.CacheSize
	worker.cache = generateRandomMessages(messagesCount)

	mockProducer.publishAssertParam = make([]producer.Message, messagesCount)
	mockProducer.publishResult = make([]error, messagesCount)
	for i, msg := range worker.cache {
		mockProducer.publishAssertParam[i] = *msg
		mockProducer.publishResult[i] = nil
	}

	assert.Equal(t, messagesCount, len(worker.cache))
	worker.Execute()
	assert.Equal(t, 0, len(worker.cache))
}

func TestBridgeWorker_Execute_flushTimeNoCacheSize(t *testing.T) {
	worker := getDefaultBridgeWorker(t)

	mockProducer := &mockProducer{t: t}
	worker.producer = mockProducer

	worker.lastFlush = time.Time{}
	messagesCount := worker.config.CacheSize - 1
	worker.cache = generateRandomMessages(messagesCount)

	mockProducer.publishAssertParam = make([]producer.Message, messagesCount)
	mockProducer.publishResult = make([]error, messagesCount)
	for i, msg := range worker.cache {
		mockProducer.publishAssertParam[i] = *msg
		mockProducer.publishResult[i] = nil
	}

	assert.Equal(t, messagesCount, len(worker.cache))
	worker.Execute()
	assert.Equal(t, 0, len(worker.cache))
}

func TestNewBridgeWorker_publishMessages(t *testing.T) {
	worker := getDefaultBridgeWorker(t)

	mockProducer := &mockProducer{t: t}
	mockStorage := &mockStorage{t: t}

	worker.producer = mockProducer
	worker.storage = mockStorage

	messagesCount := 5
	messages := generateRandomMessages(messagesCount)

	mockProducer.publishAssertParam = make([]producer.Message, messagesCount)
	mockProducer.publishResult = make([]error, messagesCount)
	for i, msg := range messages {
		mockProducer.publishAssertParam[i] = *msg
		mockProducer.publishResult[i] = nil
	}

	// let's screw some messages publishing to test fallback to storage
	mockProducer.publishResult[2] = errors.New("error for publish #2")
	mockProducer.publishResult[3] = errors.New("error for publish #3")

	mockStorage.putResult = []error{nil, errors.New("error for storage.Put() #1")}

	worker.publishMessages(messages)

	// all messages tried to be published
	assert.Equal(t, messagesCount, mockProducer.publishCalled)
	// two publish errors called storage
	assert.Equal(t, 2, mockStorage.putCalled)
	// one failed storage call returned message to cache
	assert.Equal(t, 1, len(worker.cache))
}

func TestBridgeWorker_populateCacheFromStorage(t *testing.T) {
	worker := getDefaultBridgeWorker(t)

	mockStorage := &mockStorage{t: t}
	worker.storage = mockStorage

	normalMessages := generateRandomMessages(2)
	jsonData1, _ := json.Marshal(normalMessages[0])
	jsonData2, _ := json.Marshal(normalMessages[1])

	// normal message first
	mockStorage.getResult = append(mockStorage.getResult, mockGetResult{jsonData1, nil})
	// broken json
	mockStorage.getResult = append(mockStorage.getResult, mockGetResult{[]byte("i no json"), nil})
	// problems with storage
	mockStorage.getResult = append(mockStorage.getResult, mockGetResult{nil, errors.New("some problems")})
	// second normal message
	mockStorage.getResult = append(mockStorage.getResult, mockGetResult{jsonData2, nil})

	assert.Equal(t, 0, len(worker.cache))
	worker.populateCacheFromStorage()
	assert.Equal(t, 2, len(worker.cache))
	assert.Equal(t, normalMessages, worker.cache)
}

func TestBridgeWorker_populateCacheFromStorage_maxErrors(t *testing.T) {
	worker := getDefaultBridgeWorker(t)

	mockStorage := &mockStorage{t: t}
	worker.storage = mockStorage
	worker.config.StorageMaxErrors = 2

	normalMessages := generateRandomMessages(3)
	jsonData1, _ := json.Marshal(normalMessages[0])
	jsonData2, _ := json.Marshal(normalMessages[1])
	jsonData3, _ := json.Marshal(normalMessages[2])

	// normal message first
	mockStorage.getResult = append(mockStorage.getResult, mockGetResult{jsonData1, nil})
	// broken json
	mockStorage.getResult = append(mockStorage.getResult, mockGetResult{[]byte("i no json"), nil})
	// problems with storage
	mockStorage.getResult = append(mockStorage.getResult, mockGetResult{nil, errors.New("some problems")})
	// broken json again, this resets max errors counter
	mockStorage.getResult = append(mockStorage.getResult, mockGetResult{[]byte("i no json"), nil})
	// second normal message
	mockStorage.getResult = append(mockStorage.getResult, mockGetResult{jsonData2, nil})
	// problems with storage two times, so it should break the cycle
	mockStorage.getResult = append(mockStorage.getResult, mockGetResult{nil, errors.New("some problems")})
	mockStorage.getResult = append(mockStorage.getResult, mockGetResult{nil, errors.New("some problems")})
	// third normal message, this should never be processed
	mockStorage.getResult = append(mockStorage.getResult, mockGetResult{jsonData3, nil})

	assert.Equal(t, 0, len(worker.cache))
	worker.populateCacheFromStorage()
	assert.Equal(t, 2, len(worker.cache))
	assert.Equal(t, normalMessages[:2], worker.cache)
}

func TestBridgeWorker_Close(t *testing.T) {
	worker := getDefaultBridgeWorker(t)

	errStorageClose := errors.New("some storage error")
	mockStorage := &mockStorage{t: t, closeResult: errStorageClose}
	worker.storage = mockStorage
	worker.readStorageTicker = time.NewTicker(worker.config.StorageReadTimeout)

	messagesCount := 5
	messages := generateRandomMessages(messagesCount)

	worker.cache = messages
	for i := 0; i < len(messages); i++ {
		mockStorage.putResult = append(mockStorage.putResult, nil)
	}

	assert.Equal(t, messagesCount, len(worker.cache))
	assert.Equal(t, 0, len(mockStorage.putData))

	err := worker.Close()
	assert.Error(t, err)
	assert.Equal(t, errStorageClose, err)
	assert.Equal(t, messagesCount, mockStorage.putCalled)
	assert.Equal(t, messagesCount, len(mockStorage.putData))
	for i, msg := range messages {
		var storedMsg *producer.Message
		err = json.Unmarshal(mockStorage.putData[i], &storedMsg)
		assert.Nil(t, err)
		assert.Equal(t, msg, storedMsg)
	}
}

func TestBridgeWorker_storeMessage_errors(t *testing.T) {
	worker := getDefaultBridgeWorker(t)
	mockStorage := &mockStorage{t: t, putResult: []error{nil, errors.New("some put error")}}
	worker.storage = mockStorage

	// I tried to make a json marshal error here, but no luck, it works
	msg1 := producer.NewMessage([]byte("Z̮̞̠͙͔ͅḀ̗̞͈̻̗Ḷ͙͎̯̹̞͓G̻O̭̗̮"), "🇺🇸🇷🇺🇸 🇦🇫🇦🇲🇸")
	err := worker.storeMessage(msg1)
	assert.NoError(t, err)

	msg2 := producer.NewMessage([]byte(uuid.NewV4().String()), uuid.NewV4().String())
	err = worker.storeMessage(msg2)
	assert.Error(t, err)
	assert.Equal(t, errPutToStorage, err)

	assert.Equal(t, 2, len(mockStorage.putData))

	var msg1Json *producer.Message
	err = json.Unmarshal(mockStorage.putData[0], &msg1Json)
	assert.NoError(t, err)
	assert.Equal(t, msg1, msg1Json)

	var msg2Json *producer.Message
	err = json.Unmarshal(mockStorage.putData[1], &msg2Json)
	assert.NoError(t, err)
	assert.Equal(t, msg2, msg2Json)
}
