package producer

import (
	"errors"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/hellofresh/stats-go"
	"github.com/stretchr/testify/assert"
)

type sendMessageResult struct {
	partition int32
	offset    int64
	err       error
}

type mockSyncProducer struct {
	sendMessageResult  sendMessageResult
	sendMessagesResult error
	closeResult        error

	lastSendMessageParams  *sarama.ProducerMessage
	lastSendMessagesParams []*sarama.ProducerMessage
}

func (p *mockSyncProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	p.lastSendMessageParams = msg
	return p.sendMessageResult.partition, p.sendMessageResult.offset, p.sendMessageResult.err
}

func (p *mockSyncProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	p.lastSendMessagesParams = msgs
	return p.sendMessagesResult
}

func (p *mockSyncProducer) Close() error {
	return p.closeResult
}

func TestKafkaProducer_Close(t *testing.T) {
	closeError := errors.New("Close error result")

	mockProducer := &mockSyncProducer{closeResult: closeError}
	statsClient := stats.NewStatsdStatsClient("", "")

	kafkaProducer := &KafkaProducer{mockProducer, statsClient}

	err := kafkaProducer.Close()
	assert.NotEmpty(t, err)
	assert.Equal(t, closeError, err)
}

func TestKafkaProducer_Publish(t *testing.T) {
	mockProducer := &mockSyncProducer{}
	statsClient := stats.NewStatsdStatsClient("", "")

	body := "hello message body!"
	topic := "some topic"
	msg := NewMessage([]byte(body), topic)

	kafkaProducer := &KafkaProducer{mockProducer, statsClient}

	err := kafkaProducer.Publish(*msg)
	assert.Nil(t, err)

	messageValue, err := mockProducer.lastSendMessageParams.Value.Encode()
	assert.Nil(t, err)
	assert.Equal(t, body, string(messageValue))
	assert.Equal(t, topic, mockProducer.lastSendMessageParams.Topic)
}

func TestKafkaProducer_Publish_error(t *testing.T) {
	sendMessageError := errors.New("Send message error")
	sendMessageResult := sendMessageResult{0, 0, sendMessageError}

	mockProducer := &mockSyncProducer{sendMessageResult: sendMessageResult}
	statsClient := stats.NewStatsdStatsClient("", "")

	body := "hello message body!"
	topic := "some topic"
	msg := NewMessage([]byte(body), topic)

	kafkaProducer := &KafkaProducer{mockProducer, statsClient}

	err := kafkaProducer.Publish(*msg)
	assert.NotEmpty(t, err)
	assert.Equal(t, sendMessageError, err)
}
