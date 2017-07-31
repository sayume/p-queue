package main

import (
	"time"

	pq "p-queue/model"
)

var (
	pivotTime = time.Now().UnixNano()
)

type Element struct {
	id              string
	session         string
	Priority        int       `json:"priority"`
	CreateTime      int64     `json:"startTimestamp"`
	EstimateTime    int64     `json:"estimateTime"`
	Business        string    `json:"business"`
	RequireResource *Resource `json:"requireResource"`
	Region          string    `json:"region"`
	ISP             string    `json:"isp"`
}

type Resource struct {
	Cpus      float64 `json:"cpus"`
	Mem       float64 `json:"mem"`
	Disk      float64 `json:"disk"`
	Bandwidth float64 `json:"bandwidth"`
}

func (e *Element) GetID() string {
	return e.id
}

func (e *Element) SetID(id string) {
	e.id = id
}

// Revise this method to implement custom algorithm
func (e *Element) GetScore() float64 {
	score := int64(e.Priority) * (e.CreateTime - pivotTime)
	return float64(score)
}

// Revise this method to implement custom timeout
func (e *Element) GetTimeout() int64 {
	return e.estimateExecutionTime * 2
}

func NewElement() pq.QueueElement {
	return &Element{}
}
