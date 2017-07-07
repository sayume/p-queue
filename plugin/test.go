package main

import (
	"encoding/json"
	"fmt"
	"plugin"
	"time"

	. "p-queue/model"
)

type TaskMeta struct {
	Priority   int   `json:"priority"`
	CreateTime int64 `json:"createTime"`
}

func main() {
	p, err := plugin.Open("strategy.so")
	if err != nil {
		panic(err)
	}
	f, err := p.Lookup("NewElement")
	if err != nil {
		panic(err)
	}

	var element QueueElement
	element = f.(func() QueueElement)()

	taskMeta := &TaskMeta{
		Priority:   1,
		CreateTime: time.Now().UnixNano(),j
	}
	data, err := json.Marshal(taskMeta)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal(data, element)
	if err != nil {
		fmt.Println(err)
		return
	}
	element.SetID("aaa")
	fmt.Println(element.GetID())
	fmt.Println(element.GetScore())
}
