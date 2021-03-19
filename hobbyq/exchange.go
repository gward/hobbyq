package hobbyq

type Exchange struct {
	name string
	subscriptions []Subscription
}

type Subscription struct {
	topicKey string
	queue Queue
}

func NewExchange(name string) *Exchange {
	return &Exchange{name, nil}
}
