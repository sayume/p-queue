package main

import (
	"fmt"
	"plugin"

	. "p-queue/model"
)

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

	element.SetID("aaa")
	fmt.Println(element.GetID())
	fmt.Println(element.GetScore())
}
