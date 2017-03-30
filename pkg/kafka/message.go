package kafka

import (
	"fmt"

	"github.com/satori/go.uuid"
)

// Message struct contains data for message read from RabbitMQ and ready for sending to Kafka
type Message struct {
	ID    uuid.UUID `json:"id"`
	Body  []byte    `json:"body"`
	Topic string    `json:"topic"`
}

// NewMessage initializes and instantiates new Message
func NewMessage(body []byte, topic string) *Message {
	return &Message{uuid.NewV4(), body, topic}
}

func (m Message) String() string {
	return fmt.Sprintf("{id: %s, topic: %s}", m.ID.String(), m.Topic)
}
