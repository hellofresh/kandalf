package workers

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/hellofresh/stats-go/bucket"
	"github.com/hellofresh/stats-go/client"
	log "github.com/sirupsen/logrus"

	"github.com/hellofresh/kandalf/pkg/config"
	"github.com/hellofresh/kandalf/pkg/producer"
	"github.com/hellofresh/kandalf/pkg/storage"
)

const (
	statsWorkerSection = "worker"
)

var (
	errMarshalMessage = errors.New("failed to marshal message")
	errPutToStorage   = errors.New("failed to put message to storage")
)

// BridgeWorker contains data for bridge worker that does the actual job - handles messages transfer
// from RabbitMQ to Kafka
type BridgeWorker struct {
	sync.Mutex

	config      config.WorkerConfig
	storage     storage.PersistentStorage
	producer    producer.Producer
	statsClient client.Client

	cache             []*producer.Message
	lastFlush         time.Time
	readStorageTicker *time.Ticker
}

// NewBridgeWorker creates instance of BridgeWorker
func NewBridgeWorker(config config.WorkerConfig, storage storage.PersistentStorage, producer producer.Producer, statsClient client.Client) (*BridgeWorker, error) {
	return &BridgeWorker{config: config, storage: storage, producer: producer, statsClient: statsClient}, nil
}

// Execute runs the service logic once in sync way
func (w *BridgeWorker) Execute() {
	w.Lock()
	defer w.Unlock()

	if len(w.cache) >= w.config.CacheSize || time.Now().Sub(w.lastFlush) >= w.config.CacheFlushTimeout {
		log.WithFields(log.Fields{"len": len(w.cache), "last_flush": w.lastFlush}).
			Debug("Flushing worker cache to Kafka")

		if len(w.cache) > 0 {
			// copy workers cache to local cache to avoid long locking for worker cache,
			// as all incoming messages will be waiting for network communication with kafka/storage
			messages := make([]*producer.Message, len(w.cache))
			copy(messages, w.cache)
			w.cache = []*producer.Message{}

			go w.publishMessages(messages)
		}
		w.lastFlush = time.Now()
	}
}

// Go runs the service forever in async way in go-routine
func (w *BridgeWorker) Go(ctx context.Context) {
	w.readStorageTicker = time.NewTicker(w.config.StorageReadTimeout)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-w.readStorageTicker.C:
				w.populateCacheFromStorage()
			default:
				w.Execute()
			}

			// Prevent CPU overload
			log.WithField("timeout", w.config.CycleTimeout.String()).Debug("Bridge worker is going to sleep for a while")
			time.Sleep(w.config.CycleTimeout)
		}
	}()
}

// Close closes worker resources
func (w *BridgeWorker) Close() error {
	log.Info("Closing bridge worker, will handle storage close either")

	// stop storage reader
	w.readStorageTicker.Stop()

	// lock cache and save all unhandled messages to storage for further processing
	// do not unlock cache anymore as we're closing everything
	w.Lock()
	log.WithField("len", len(w.cache)).Info("Storing unhandled messages to storage")
	for _, msg := range w.cache {
		// do not handle errors here as there is nothing we can do with errors at this point
		w.storeMessage(msg)
	}

	return w.storage.Close()
}

// MessageHandler is a handler function for new messages from AMQP
func (w *BridgeWorker) MessageHandler(body []byte, pipe config.Pipe) error {
	return w.cacheMessage(producer.NewMessage(body, pipe.KafkaTopic))
}

func (w *BridgeWorker) cacheMessage(msg *producer.Message) error {
	w.Lock()
	defer w.Unlock()

	w.cache = append(w.cache, msg)

	operation := bucket.NewMetricOperation("cache", "add", msg.Topic)
	w.statsClient.TrackOperation(statsWorkerSection, operation, nil, true)

	return nil
}

func (w *BridgeWorker) populateCacheFromStorage() {
	var errorsCount int

	log.Debug("Populating cache from storage")
	for {
		if errorsCount >= w.config.StorageMaxErrors {
			log.WithField("errors_count", errorsCount).
				Error("Got several errors in a row while reading from storage, stopping reading")
			break
		}

		operation := bucket.NewMetricOperation("storage", "get")
		storageMsg, err := w.storage.Get()
		if err != nil {
			if err == storage.ErrStorageIsEmpty {
				break
			}
			log.WithError(err).Error("Failed to read message from persistent storage")
			w.statsClient.TrackOperation(statsWorkerSection, operation, nil, false)
			errorsCount++
			continue
		}
		w.statsClient.TrackOperation(statsWorkerSection, operation, nil, true)
		errorsCount = 0

		operation = bucket.NewMetricOperation("storage", "unmarshal")
		var msg *producer.Message
		err = json.Unmarshal(storageMsg, &msg)
		if err != nil {
			log.WithError(err).Error("Failed to unmarshal message from persistent storage")
			w.statsClient.TrackOperation(statsWorkerSection, operation, nil, false)
			continue
		}
		w.statsClient.TrackOperation(statsWorkerSection, operation, nil, true)

		w.cacheMessage(msg)
	}
}

func (w *BridgeWorker) publishMessages(messages []*producer.Message) {
	for _, msg := range messages {
		err := w.producer.Publish(*msg)
		if err != nil {
			log.WithError(err).WithField("msg", msg.String()).
				Warning("Failed to publish messages to Kafka, moving to storage")

			if err = w.storeMessage(msg); err != nil {
				if err == errMarshalMessage {
					continue
				} else if err == errPutToStorage {
					w.cacheMessage(msg)
				} else {
					log.WithError(err).WithField("msg", msg.String()).
						Error("Unhandled storage error")
				}
			}
		}
	}
}

func (w *BridgeWorker) storeMessage(msg *producer.Message) error {
	data, err := json.Marshal(msg)

	operation := bucket.NewMetricOperation("storage", "marshal")
	w.statsClient.TrackOperation(statsWorkerSection, operation, nil, err == nil)

	if err != nil {
		log.WithError(err).WithField("msg", msg.String()).
			Error("Failed to marshal message")
		return errMarshalMessage
	}

	err = w.storage.Put(data)

	operation = bucket.NewMetricOperation("storage", "set")
	w.statsClient.TrackOperation(statsWorkerSection, operation, nil, err == nil)

	if err != nil {
		log.WithError(err).WithField("msg", msg.String()).
			Error("Failed to put message to storage, returning to cache")
		return errPutToStorage
	}

	return nil
}
