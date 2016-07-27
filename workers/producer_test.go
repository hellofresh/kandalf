package workers

import (
	"sort"
	"testing"

	"kandalf/pipes"
)

func Test_getTopic_RoutingKeysExact(t *testing.T) {
	msg := internalMessage{
		RoutingKeys: []string{
			"customers_us",
		},
	}

	topic := getTopic(msg, getPipes())

	if len(topic) == 0 || topic != "customers_us" {
		t.Error("Expected `customers_us`, got ", topic)
	}
}

func Test_getTopic_RoutingKeys(t *testing.T) {
	msg := internalMessage{
		RoutingKeys: []string{
			"customers_ca",
		},
	}

	topic := getTopic(msg, getPipes())

	if len(topic) == 0 || topic != "customers" {
		t.Error("Expected `customers`, got ", topic)
	}
}

func Test_getTopic_RoutedQueue(t *testing.T) {
	msg := internalMessage{
		RoutedQueues: []string{
			"order.new",
		},
	}

	topic := getTopic(msg, getPipes())

	if len(topic) == 0 || topic != "orders" {
		t.Error("Expected `orders`, got ", topic)
	}
}

func Test_getTopic_Empty(t *testing.T) {
	msg := internalMessage{
		RoutingKeys: []string{
			"key",
		},
		RoutedQueues: []string{
			"queue",
		},
	}

	topic := getTopic(msg, getPipes())

	if len(topic) > 0 {
		t.Error("Expected empty topic, got ", topic)
	}
}

func getPipes() (p pipes.PipesList) {
	p = pipes.PipesList{
		{
			Topic:         "customers",
			RoutingKey:    "customers_*",
			Priority:      2,
			HasRoutingKey: true,
		},
		{
			Topic:          "orders",
			RoutedQueue:    "order*",
			Priority:       2,
			HasRoutedQueue: true,
		},
		{
			Topic:         "customers_us",
			RoutingKey:    "customers_us*",
			Priority:      3,
			HasRoutingKey: true,
		},
	}

	sort.Sort(p)

	return p
}
