package workers

import (
	"encoding/json"
	"sync"
	"time"

	"gopkg.in/redis.v3"

	"../config"
	"../logger"
)

type cbHandleMessage func(Message) error

type Message struct {
	Exchange   string          `json:"exchange"`
	RoutingKey string          `json:"routing_key"`
	Body       json.RawMessage `json:"body"`
}

type Queue struct {
	flushTimeout   time.Duration
	handleMessage  cbHandleMessage
	isWorking      bool
	maxSize        int
	messages       []Message
	mutex          *sync.Mutex
	rd             *redis.Client
	rdFlushTimeout time.Duration
	rdKeyName      string
}

func NewQueue(handler cbHandleMessage) (*Queue, error) {
	var err error

	rdKeyName, err := config.Instance().String("queue.redis.key_name")
	if err != nil {
		return nil, err
	}

	rdAddr, err := config.Instance().String("queue.redis.address")
	if err != nil {
		return nil, err
	}

	rdClient := redis.NewClient(&redis.Options{
		Addr: rdAddr,
	})

	_, err = rdClient.Ping().Result()
	if err != nil {
		return nil, err
	}

	rdFlushTimeout, err := config.Instance().Int("queue.redis.flush_timeout")
	if err != nil {
		rdFlushTimeout = 10
	}

	flushTimeout, err := config.Instance().Int("queue.flush_timeout")
	if err != nil {
		flushTimeout = 5
	}

	maxSize, err := config.Instance().Int("queue.max_size")
	if err != nil {
		maxSize = 10
	}

	return &Queue{
		flushTimeout:   flushTimeout * time.Second,
		handleMessage:  handler,
		maxSize:        maxSize,
		messages:       make([]Message, 0),
		mutex:          &sync.Mutex{},
		rd:             rdClient,
		rdFlushTimeout: rdFlushTimeout * time.Second,
		rdKeyName:      rdKeyName,
	}, nil
}

// Main working cycle
func (q *Queue) Run(wg *sync.WaitGroup, die chan bool) {
	defer wg.Done()

	q.isWorking = true

	// Periodically flush the queue to the kafka
	go func() {
		for {
			select {
			case <-die:
				q.isWorking = false
				// Before exiting the working cycle, try to flush the messages to kafka
				q.handleMessages()
				return
			default:
			}

			if len(q.messages) >= q.maxSize {
				q.handleMessages()
			} else {
				time.Sleep(q.flushTimeout)
			}
		}
	}()

	// And move messages from redis back to normal queue
	go q.flushRedis()
}

// Adds message to the internal queue which will be flushed to kafka later
func (q *Queue) Add(msg Message) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	append(q.messages, msg)
}

// Tries to send messages to the kafka
// If message failed, it will be stored in redis
func (q *Queue) handleMessages() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	var (
		err            error
		failedMessages []Message = make([]Message, 0)
	)

	for msg := range q.messages {
		err = q.handleMessage(msg)
		if err == nil {
			continue
		}

		// If it is not possible to store message even in redis,
		// we'll put the message into memory and process it later
		err = q.storeInRedis(msg)
		if err != nil {
			failedMessages = append(failedMessages, msg)
		}
	}

	q.messages = failedMessages
}

// Stores the failed message in redis
func (q *Queue) storeInRedis(msg Message) error {
	strMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return q.rd.LPush(q.rdKeyName, strMsg).Err()
}

// Periodically move messages from Redis back to the main queue
func (q *Queue) flushRedis() {
	var (
		err      error
		bytesMsg []byte
		msg      Message
	)

	for q.isWorking {
		for {
			bytesMsg, err = q.rd.LPop(q.rdKeyName).Bytes()
			if err != nil || len(bytesMsg) == 0 {
				break
			}

			err = json.Unmarshal(bytesMsg, &msg)
			if err != nil {
				logger.Instance().
					WithError(err).
					Warning("Unable to unmarshal JSON from redis")
				continue
			}

			q.Add(msg)
		}

		time.Sleep(q.rdFlushTimeout)
	}

	// Close connection to redis
	err = q.rd.Close()
	if err != nil {
		logger.Instance().
			WithError(err).
			Warning("An error occurred while closing connection to redis")
	}
}
