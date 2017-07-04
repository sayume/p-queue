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
	Priority   int   `json:"priority"`
	CreateTime int64 `json:"createTime"`
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
	score := int64(e.Priority) * (e.CreateTime - pivotTime)
	return float64(score)
}

func NewElement() pq.QueueElement {
	return &Element{}
}
