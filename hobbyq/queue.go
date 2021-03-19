package hobbyq

type Queue struct {
	name string
}

func NewQueue(name string) *Queue {
	return &Queue{name}
}
