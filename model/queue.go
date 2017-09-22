package pq

// Need to fix
type PQueue interface {
	Push(element QueueElement) error
	Pop(element QueueElement) error
	Ack(session string) (chan bool, error)
}
