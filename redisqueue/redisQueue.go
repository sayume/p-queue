package redisqueue

import (
	"errors"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	uuid "github.com/pborman/uuid"

	pq "p-queue/model"
)

var (
	errRedis         = errors.New("Fail to manipulate redis store.")
	errParseSession  = errors.New("Invalid session.")
	errParse         = errors.New("Parse data type error.")
	errQueueFull     = errors.New("Queue is full.")
	errQueueIsEmpty  = errors.New("Queue is empty.")
	cancelChannelMap = make(map[string](chan bool))
)

type RedisQueueConfig struct {
	Addr      string
	ID        string
	MaxLength int
}

type RedisQueue struct {
	config     *RedisQueueConfig
	client     *redis.Client
	id         string
	timeoutVal time.Duration
	retryTimes int
	length     int
}

func NewRedisQueue(config *RedisQueueConfig) *RedisQueue {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: "",
		DB:       0,
	})
	id := uuid.NewUUID().String()
	if config.ID != "" {
		id = config.ID
	}
	queue := &RedisQueue{
		config:     config,
		client:     client,
		id:         id,
		timeoutVal: time.Duration(30) * time.Second,
		retryTimes: 3,
		length:     0,
	}
	return queue
}

func (r *RedisQueue) buildElementPrefix() string {
	return r.id + ":element"
}

func (r *RedisQueue) buildQueuePrefix() string {
	return r.id + ":queue"
}

func buildSession(id string) string {
	return id + "|session"
}

func (r *RedisQueue) isQueueFull() bool {
	return r.length >= r.config.MaxLength
}

func extractIDFromSession(session string) (string, error) {
	array := strings.Split(session, "|")
	if len(array) != 2 {
		return "", errParseSession
	}
	id := array[0]
	return id, nil
}

func (r *RedisQueue) collectElement(e pq.QueueElement, c chan bool, timeoutChan chan bool) {
	defer close(timeoutChan)
	timeout := e.GetTimeout()
	if r.timeoutVal.Nanoseconds() > e.GetTimeout() {
		timeout = r.timeoutVal.Nanoseconds()
	}
	select {
	case <-time.After(time.Duration(timeout) * time.Nanosecond):
		// Close channel
		close(c)
		session := buildSession(e.GetID())
		if cancelChannelMap[session] != nil {
			delete(cancelChannelMap, session)
		}
		// Re-push element to the queue
		result1 := r.client.HGet(r.buildElementPrefix(), session)
		err := result1.Err()
		if err != nil {
			log.Error(err)
			// Element already acked by other client, just ignore.
			return
		}
		times, err := result1.Int64()
		if err != nil {
			log.Error(err)
			return
		}
		if times > int64(r.retryTimes) {
			result2 := r.client.HDel(r.buildElementPrefix(), session)
			err = result2.Err()
			if err != nil {
				log.Error(err)
				return
			}
			log.WithField("elementID", e.GetID()).Info("Element is out of retry limit, delete from queue.")
			timeoutChan <- true
			return
		}
		// Put the element back to the queue.
		r.Push(e)
	case <-c:
		return
	}
}

func (r *RedisQueue) Push(e pq.QueueElement) error {
	if r.isQueueFull() {
		return errQueueFull
	}
	result := r.client.ZAdd(r.buildQueuePrefix(), redis.Z{
		Score:  e.GetScore(),
		Member: e.GetID(),
	})
	err := result.Err()
	if err != nil {
		log.Error(err)
		return errRedis
	}
	r.length++
	return nil
}

func (r *RedisQueue) Pop(element pq.QueueElement) (chan bool, error) {
	result1 := r.client.ZRange(r.buildQueuePrefix(), 0, 0)
	err := result1.Err()
	if err != nil {
		log.Error(err)
		return nil, errRedis
	}
	values := result1.Val()
	if len(values) == 0 {
		return nil, errQueueIsEmpty
	}
	id := values[0]
	element.SetID(id)

	// Use transaction to make sure every pop is an atomic exec.
	session := buildSession(id)
	pipe := r.client.TxPipeline()
	pipe.ZRemRangeByRank(r.buildQueuePrefix(), 0, 0)
	pipe.HSetNX(r.buildElementPrefix(), session, 0)
	pipe.Expire(r.buildElementPrefix(), 24*time.Hour)
	pipe.HIncrBy(r.buildElementPrefix(), session, 1)
	_, err = pipe.Exec()
	if err != nil {
		log.Error(err)
		return nil, errRedis
	}
	cancelChannel := make(chan bool, 1)
	timeoutChannel := make(chan bool, 1)
	go r.collectElement(element, cancelChannel, timeoutChannel)
	cancelChannelMap[session] = cancelChannel

	return timeoutChannel, nil
}

func (r *RedisQueue) Ack(id string) error {
	session := buildSession(id)
	result := r.client.HDel(r.buildElementPrefix(), session)
	err := result.Err()
	if err != nil {
		log.Error(err)
		return errRedis
	}
	if cancelChannelMap[session] != nil {
		cancelChannelMap[session] <- true
		delete(cancelChannelMap, session)
	}
	r.length--
	return nil
}

func (r *RedisQueue) GetRetryTimes(id string) (int, error) {
	session := buildSession(id)
	result := r.client.HGet(r.buildElementPrefix(), session)
	err := result.Err()
	if err != nil {
		log.Error(err)
		return 0, errRedis
	}
	times, err := result.Int64()
	if err != nil {
		log.Error(err)
		return 0, errParse
	}
	return int(times), nil
}

func (r *RedisQueue) GetQueueLength() int {
	return r.length
}
