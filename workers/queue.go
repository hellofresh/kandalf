package workers

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/hellofresh/kandalf/pkg/config"
	"github.com/hellofresh/stats-go"
	"gopkg.in/redis.v3"
)

const (
	statsQueueSection = "queue"
)

type internalMessage struct {
	Body  json.RawMessage `json:"body"`
	Topic string          `json:"topic"`
}

type internalQueue struct {
	isWorking   bool
	messages    []internalMessage
	mutex       sync.Mutex
	producer    *internalProducer
	rd          *redis.Client
	statsClient stats.StatsClient
	redisConfig config.RedisConfig
}

// Returns new instance of queue
func newInternalQueue(globalConfig config.GlobalConfig, statsClient stats.StatsClient) (*internalQueue, error) {
	var err error

	rdClient := redis.NewClient(&redis.Options{Addr: globalConfig.Redis.Address})

	_, err = rdClient.Ping().Result()
	if err != nil {
		return nil, err
	}

	producer, err := newInternalProducer(globalConfig.Kafka, statsClient)
	if err != nil {
		return nil, fmt.Errorf("An error occured while instantiating producer: %v", err)
	}

	return &internalQueue{
		messages:    make([]internalMessage, 0),
		producer:    producer,
		rd:          rdClient,
		statsClient: statsClient,
		redisConfig: globalConfig.Redis,
	}, nil
}

// Main working cycle
func (q *internalQueue) run(wg *sync.WaitGroup, die chan bool) {
	defer wg.Done()

	q.isWorking = true

	// Periodically flush the queue to the kafka
	go func() {
		var (
			lastFlush  time.Time = time.Now()
			nbMessages int
			oneSec     time.Duration = time.Second
		)

		for {
			select {
			case <-die:
				q.isWorking = false

				// Before exiting the working cycle, try to flush the messages to kafka
				q.handleMessages()
				return
			default:
			}

			nbMessages = len(q.messages)

			if (nbMessages >= q.redisConfig.PoolSize) || (nbMessages > 0 && time.Now().Sub(lastFlush) > q.redisConfig.FlushTimeout.Duration) {
				q.handleMessages()

				lastFlush = time.Now()
			} else {
				time.Sleep(oneSec)
			}
		}
	}()

	// And move messages from redis back to normal queue
	go q.flushRedis()
}

// Adds message to the internal queue which will be flushed to kafka later
func (q *internalQueue) add(msg internalMessage) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.messages = append(q.messages, msg)

	operation := stats.MetricOperation{"internal-queue", "new-messages", stats.MetricEmptyPlaceholder}
	q.statsClient.TrackOperation(statsQueueSection, operation, nil, true)

	log.WithField("topic", msg.Topic).Debug("Added message to internal queue")
}

// Tries to send messages to the kafka
// If message failed, it will be stored in redis
func (q *internalQueue) handleMessages() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	var (
		err            error
		failedMessages []internalMessage = make([]internalMessage, 0)
	)

	log.Debug("Start sending messages to kafka")

	for _, msg := range q.messages {
		err = q.producer.handleMessage(msg)
		if err == nil {
			continue
		} else {
			if strings.Contains(err.Error(), "topic or partition that does not exist") {
				continue
			}

			log.WithError(err).Warning("Unable to send message to kafka")
		}

		err = q.storeInRedis(msg)

		// If it is not possible to store message even in redis,
		// we'll put the message into memory and process it later
		operation := stats.MetricOperation{"redis", "messages", stats.MetricEmptyPlaceholder}
		if err != nil {
			failedMessages = append(failedMessages, msg)

			q.statsClient.TrackOperation(statsQueueSection, operation, nil, false)
			log.WithError(err).Warning("Unable to store message in Redis")
		} else {
			q.statsClient.TrackOperation(statsQueueSection, operation, nil, true)
			log.Debug("Successfully stored message in Redis")
		}
	}

	q.messages = failedMessages
}

// Stores the failed message in redis
func (q *internalQueue) storeInRedis(msg internalMessage) error {
	strMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return q.rd.LPush(q.redisConfig.KeyName, string(strMsg)).Err()
}

// Periodically move messages from Redis back to the main queue
func (q *internalQueue) flushRedis() {
	var (
		err      error
		bytesMsg []byte
		msg      internalMessage
	)

	for q.isWorking {
		for {
			bytesMsg, err = q.rd.LPop(q.redisConfig.KeyName).Bytes()
			if err != nil || len(bytesMsg) == 0 {
				break
			}

			err = json.Unmarshal(bytesMsg, &msg)
			if err != nil {
				log.WithError(err).Warning("Unable to unmarshal JSON from redis")
				continue
			}

			q.add(msg)
		}

		time.Sleep(q.redisConfig.ReadTimeout.Duration)
	}

	// Close connection to redis
	err = q.rd.Close()
	if err != nil {
		log.WithError(err).Warning("An error occurred while closing connection to redis")
	}
}
