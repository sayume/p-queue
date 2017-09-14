package pq

type QueueElement interface {
	GetScore() float64
	GetTimeout() int64
	GetID() string
	SetID(id string)
	SetTimeout(timeout int64)
}
