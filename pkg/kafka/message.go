package kafka

import (
	"fmt"

	"github.com/satori/go.uuid"
)

type Message struct {
	ID    uuid.UUID `json:"id"`
	Body  []byte    `json:"body"`
	Topic string    `json:"topic"`
}

func NewMessage(body []byte, topic string) *Message {
	return &Message{uuid.NewV4(), body, topic}
}

func (m Message) String() string {
	return fmt.Sprintf("{id: %s, topic: %s}", m.ID.String(), m.Topic)
}
