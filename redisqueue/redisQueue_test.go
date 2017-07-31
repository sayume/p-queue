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

func (e Element) GetTimeout() int64 {
	return e.estimateExecutionTime * 2
}

func TestRedisQueue(t *testing.T) {
	config := &RedisQueueConfig{
		Addr: "127.0.0.1:6379",
	}
	queue := NewRedisQueue(config)

	estimate := 1 * time.Second
	Convey("Push two elements in queue.", t, func() {
		var element1 QueueElement
		element1 = &Element{
			id:                    "aaa",
			priority:              1,
			createTime:            time.Now().UnixNano(),
			estimateExecutionTime: estimate.Nanoseconds(),
		}
		var element2 QueueElement
		element2 = &Element{
			id:                    "bbb",
			priority:              1,
			createTime:            time.Now().UnixNano(),
			estimateExecutionTime: estimate.Nanoseconds(),
		}

		err := queue.Push(element1)
		So(err, ShouldBeNil)
		err = queue.Push(element2)
		So(err, ShouldBeNil)
	})

	Convey("Pop an element from queue and ack", t, func() {
		var element QueueElement
		element = &Element{}
		err := queue.Pop(element)
		So(err, ShouldBeNil)
		So(element.GetID(), ShouldEqual, "aaa")
		err = queue.Ack(element.GetID())
		So(err, ShouldBeNil)
	})

	Convey("An not acked element should be timeout", t, func() {
		var element1 QueueElement
		element1 = &Element{}
		err := queue.Pop(element1)
		So(err, ShouldBeNil)
		So(element1.GetID(), ShouldEqual, "bbb")
		var element2 QueueElement
		element2 = &Element{}
		err = queue.Pop(element2)
		So(err, ShouldBeNil)
		So(element2.GetID(), ShouldEqual, "")

		<-time.After(35 * time.Second)
		var element3 QueueElement
		element3 = &Element{}
		err = queue.Pop(element3)
		So(err, ShouldBeNil)
		So(element3.GetID(), ShouldEqual, "bbb")
		err = queue.Ack(element3.GetID())
		So(err, ShouldBeNil)
	})
}
