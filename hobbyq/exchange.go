package hobbyq

import (
	"encoding/json"
)

type Exchange struct {
	name string
	subscriptions []Subscription
}

type Subscription struct {
	topicKey string
	queue Queue
}

func NewExchange(name string) *Exchange {
	return &Exchange{
		name,
		make([]Subscription, 0),
	}
}

func (exchange *Exchange) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string] interface{}{
		"name": exchange.name,
		"subscriptions": exchange.subscriptions,
	})
}
