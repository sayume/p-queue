package redisqueue

import (
	"testing"
	"time"

	. "p-queue/model"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	pivotTime = time.Now().UnixNano()
)

type Element struct {
	id                    string
	session               string
	priority              int
	createTime            int64
	estimateExecutionTime int64
	timeout               int64
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
	// To normalize the create time
	createTime := e.createTime - pivotTime

	return float64(e.priority) * float64(createTime)
}

func (e *Element) GetTimeout() int64 {
	return e.timeout
}

func (e *Element) SetTimeout(timeout int64) {
	e.timeout = timeout
}

func TestRedisQueue(t *testing.T) {
	config := &RedisQueueConfig{
		Addr:      "127.0.0.1:6379",
		ID:        "test",
		MaxLength: 1000,
	}
	queue := NewRedisQueue(config)

	estimate := 1 * time.Second
	Convey("Push two elements in queue.", t, func() {
		var element1 QueueElement
		element1 = &Element{
			id:                    "aaa",
			session:               "",
			priority:              1,
			createTime:            time.Now().UnixNano(),
			estimateExecutionTime: estimate.Nanoseconds(),
			timeout:               int64(1 * 1000 * 1000 * 1000),
		}
		var element2 QueueElement
		element2 = &Element{
			id:                    "bbb",
			session:               "",
			priority:              1,
			createTime:            time.Now().UnixNano(),
			estimateExecutionTime: estimate.Nanoseconds(),
			timeout:               int64(1 * 1000 * 1000 * 1000),
		}

		err := queue.Push(element1)
		So(err, ShouldBeNil)
		err = queue.Push(element2)
		So(err, ShouldBeNil)
	})

	Convey("Pop an element from queue and ack", t, func() {
		var element QueueElement
		element = &Element{}
		_, err := queue.Pop(element)
		So(err, ShouldBeNil)
		So(element.GetID(), ShouldEqual, "aaa")
		err = queue.Ack(element.GetID())
		So(err, ShouldBeNil)
	})

	Convey("An not acked element should be timeout", t, func() {
		var element2 QueueElement
		element2 = &Element{}
		ti1, err := queue.Pop(element2)
		So(err, ShouldBeNil)
		times1 := <-ti1
		t.Log(times1)
		t.Log("Element is timeout at the first time.")
		var element3 QueueElement
		element3 = &Element{}
		ti2, err := queue.Pop(element3)
		So(err, ShouldBeNil)
		So(element3.GetID(), ShouldEqual, "bbb")
		times2 := <-ti2
		t.Log(times2)
		t.Log("Element is timeout at the second time.")
		var element4 QueueElement
		element4 = &Element{}
		ti3, err := queue.Pop(element4)
		So(err, ShouldBeNil)
		So(element4.GetID(), ShouldEqual, "bbb")
		times3 := <-ti3
		t.Log(times3)
		t.Log("Element is timeout at the third time.")
		var element5 QueueElement
		element5 = &Element{}
		_, err = queue.Pop(element5)
		t.Log(element5.GetID())
		So(err, ShouldNotBeNil)
	})
}
