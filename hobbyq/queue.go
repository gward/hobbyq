package hobbyq

import (
	"encoding/json"
)

type Queue struct {
	name string
}

func NewQueue(name string) *Queue {
	return &Queue{name}
}

func (queue *Queue) String() string {
	return "<Queue: " + queue.name + ">"
}

func (queue *Queue) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string] interface{}{
		"name": queue.name,
	})
}
