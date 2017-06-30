package main

import (
	"time"

	pq "p-queue/model"
)

var (
	pivotTime = time.Now().UnixNano()
)

type Element struct {
	id         string
	session    string
	Priority   int
	CreateTime int64
}

func (e *Element) GetID() string {
	return e.id
}

func (e *Element) SetID(id string) {
	e.id = id
}

func (e *Element) GetSession() string {
	return e.session
}

func (e *Element) SetSession(session string) {
	e.session = session
}

func (e *Element) GetScore() float64 {
	return float64(1000)
}

func NewElement() pq.QueueElement {
	return &Element{}
}
