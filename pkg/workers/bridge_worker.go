package workers

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/hellofresh/kandalf/pkg/config"
	"github.com/hellofresh/kandalf/pkg/kafka"
	"github.com/hellofresh/kandalf/pkg/storage"
	"github.com/hellofresh/stats-go"
)

const (
	statsWorkerSection = "worker"
)

var (
	errMarshalMessage = errors.New("Failed to marshal message")
	errPutToStorage   = errors.New("Failed to put message to storage")
)

// BridgeWorker contains data for bridge worker that does the actual job - handles messages transfer
// from RabbitMQ to Kafka
type BridgeWorker struct {
	sync.Mutex

	config      config.WorkerConfig
	storage     storage.PersistentStorage
	producer    *kafka.Producer
	statsClient stats.StatsClient

	cache             []*kafka.Message
	lastFlush         time.Time
	readStorageTicker *time.Ticker
}

// NewBridgeWorker creates instance of BridgeWorker
func NewBridgeWorker(config config.WorkerConfig, storage storage.PersistentStorage, producer *kafka.Producer, statsClient stats.StatsClient) (*BridgeWorker, error) {
	return &BridgeWorker{config: config, storage: storage, producer: producer, statsClient: statsClient}, nil
}

// Execute runs the service logic once in sync way
func (w *BridgeWorker) Execute() {
	var messages []*kafka.Message

	w.Lock()
	defer w.Unlock()
	if len(w.cache) >= w.config.CacheSize || time.Now().Sub(w.lastFlush) >= w.config.CacheFlushTimeout {
		log.WithFields(log.Fields{"len": len(w.cache), "last_flush": w.lastFlush}).
			Debug("Flushing worker cache to Kafka")

		if len(w.cache) > 0 {
			// copy workers cache to local cache to avoid long locking for worker cache,
			// as all incoming messages will be waiting for network communication with kafka/storage
			copy(messages, w.cache)
			w.cache = []*kafka.Message{}
			go w.publishMessages(messages)
		}
		w.lastFlush = time.Now()
	}
}

// Go runs the service forever in async way in go-routine
func (w *BridgeWorker) Go(interrupt chan bool) {
	w.readStorageTicker = time.NewTicker(w.config.StorageReadTimeout)

	go func() {
		for {
			select {
			case <-interrupt:
				return
			case <-w.readStorageTicker.C:
				w.populateCacheFromStorage()
			default:
				w.Execute()
			}

			// Prevent CPU overload
			log.WithField("timeout", w.config.CycleTimeout).Debug("Bridge worker is going to sleep for a while")
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
	return w.cacheMessage(kafka.NewMessage(body, pipe.KafkaTopic))
}

func (w *BridgeWorker) cacheMessage(msg *kafka.Message) error {
	w.Lock()
	defer w.Unlock()

	w.cache = append(w.cache, msg)

	operation := stats.MetricOperation{"cache", "add", msg.Topic}
	w.statsClient.TrackOperation(statsWorkerSection, operation, nil, true)

	return nil
}

func (w *BridgeWorker) populateCacheFromStorage() {
	var (
		msg         kafka.Message
		errorsCount int
	)

	for {
		if errorsCount >= w.config.StorageMaxErrors {
			log.WithField("errors_count", errorsCount).
				Error("Got several errors in a row while reading from storage, stoppong reading")
			break
		}

		operation := stats.MetricOperation{"storage", "get", stats.MetricEmptyPlaceholder}
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

		operation = stats.MetricOperation{"storage", "unmarshal", stats.MetricEmptyPlaceholder}
		err = json.Unmarshal(storageMsg, msg)
		if err != nil {
			log.WithError(err).Error("Failed to unmarshal message from persistent storage")
			w.statsClient.TrackOperation(statsWorkerSection, operation, nil, false)
			continue
		}
		w.statsClient.TrackOperation(statsWorkerSection, operation, nil, true)

		w.cacheMessage(&msg)
	}
}

func (w *BridgeWorker) publishMessages(messages []*kafka.Message) {
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

func (w *BridgeWorker) storeMessage(msg *kafka.Message) error {
	data, err := json.Marshal(msg)

	operation := stats.MetricOperation{"storage", "marshal", stats.MetricEmptyPlaceholder}
	w.statsClient.TrackOperation(statsWorkerSection, operation, nil, err == nil)

	if err != nil {
		log.WithError(err).WithField("msg", msg.String()).
			Error("Failed to marshal message")
		return errMarshalMessage
	}

	err = w.storage.Put(data)

	operation = stats.MetricOperation{"storage", "set", stats.MetricEmptyPlaceholder}
	w.statsClient.TrackOperation(statsWorkerSection, operation, nil, err == nil)

	if err != nil {
		log.WithError(err).WithField("msg", msg.String()).
			Error("Failed to put message to storage, returning to cache")
		return errPutToStorage
	}

	return nil
}
