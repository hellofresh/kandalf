package producer

import (
	"github.com/Shopify/sarama"
	"github.com/hellofresh/kandalf/pkg/config"
	"github.com/hellofresh/stats-go"
	"github.com/hellofresh/stats-go/bucket"
	log "github.com/sirupsen/logrus"
)

const (
	statsKafkaSection = "kafka"
)

// KafkaProducer is a Producer implementation for publishing messages to Kafka
type KafkaProducer struct {
	kafkaClient sarama.SyncProducer
	statsClient stats.Client
}

// NewKafkaProducer instantiates and establishes new Kafka connection
func NewKafkaProducer(kafkaConfig config.KafkaConfig, statsClient stats.Client) (Producer, error) {
	cnf := sarama.NewConfig()
	cnf.Producer.RequiredAcks = sarama.WaitForAll
	cnf.Producer.Retry.Max = kafkaConfig.MaxRetry
	// Producer.Return.Successes must be true to be used in a SyncProducer
	cnf.Producer.Return.Successes = true

	client, err := sarama.NewSyncProducer(kafkaConfig.Brokers, cnf)
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{kafkaClient: client, statsClient: statsClient}, nil
}

// Close closes Kafka connection
func (p *KafkaProducer) Close() error {
	return p.kafkaClient.Close()
}

// Publish publishes message to Kafka
func (p *KafkaProducer) Publish(msg Message) error {
	_, _, err := p.kafkaClient.SendMessage(&sarama.ProducerMessage{
		Topic: msg.Topic,
		Value: sarama.ByteEncoder(msg.Body),
	})

	if err == nil {
		log.WithField("msg", msg.String()).Debug("Successfully sent message to kafka")
	} else {
		log.WithError(err).WithField("msg", msg.String()).Error("Failed to publish message to kafka")
	}
	operation := bucket.MetricOperation{"publish", msg.Topic}
	p.statsClient.TrackOperation(statsKafkaSection, operation, nil, err == nil)

	return err
}
