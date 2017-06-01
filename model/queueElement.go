package pq

type QueueElement interface {
	GetScore() float64
	GetID() string
	SetID(id string)
	GetSession() string
	SetSession(session string)
}
