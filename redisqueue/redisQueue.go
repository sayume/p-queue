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
	cancelChannelMap = make(map[string](chan bool))
)

type RedisQueueConfig struct {
	Addr string
}

type RedisQueue struct {
	client     *redis.Client
	id         string
	timeoutVal time.Duration
	retryTimes int
}

func NewRedisQueue(config *RedisQueueConfig) *RedisQueue {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: "",
		DB:       0,
	})
	id := uuid.NewUUID().String()
	return &RedisQueue{
		client:     client,
		id:         id,
		timeoutVal: time.Duration(30) * time.Second,
		retryTimes: 3,
	}
}

func (r *RedisQueue) buildElementPrefix() string {
	return r.id + ":element"
}

func (r *RedisQueue) buildQueuePrefix() string {
	return r.id + ":queue"
}

func buildSession(id string) string {
	//return id + "|" + strconv.Itoa(int(time.Now().UnixNano()))
	return id + "|session"
}

func extractIDFromSession(session string) (string, error) {
	array := strings.Split(session, "|")
	if len(array) != 2 {
		return "", errParseSession
	}
	id := array[0]
	return id, nil
}

func (r *RedisQueue) collectElement(e pq.QueueElement, c chan bool) {
	select {
	case <-time.After(r.timeoutVal):
		result1 := r.client.HGet(r.buildElementPrefix(), e.GetSession())
		err := result1.Err()
		if err != nil {
			log.Error(err)
			return
		}
		times, err := result1.Int64()
		if err != nil {
			log.Error(err)
			return
		}
		if times >= 3 {
			log.WithField("elementID", e.GetID()).Info("Element is out of retry limit, delete from queue.")
			return
		}
		// Put the element back to the queue.
		r.Push(e)
	case <-c:
		return
	}
}

func (r *RedisQueue) Push(e pq.QueueElement) error {
	log.Debug(e.GetScore())
	log.Debug(e.GetID())
	result := r.client.ZAdd(r.buildQueuePrefix(), redis.Z{
		Score:  e.GetScore(),
		Member: e.GetID(),
	})
	err := result.Err()
	if err != nil {
		log.Error(err)
		return errRedis
	}
	return nil
}

func (r *RedisQueue) Pop(element pq.QueueElement) error {
	result1 := r.client.ZRange(r.buildQueuePrefix(), 0, 0)
	err := result1.Err()
	if err != nil {
		log.Error(err)
		return errRedis
	}
	values := result1.Val()
	if len(values) == 0 {
		return nil
	}
	id := values[0]
	element.SetID(id)

	// Use transaction to make sure every pop is an atomic exec.
	session := buildSession(id)
	element.SetSession(session)
	pipe := r.client.TxPipeline()
	pipe.ZRemRangeByRank(r.buildQueuePrefix(), 0, 0)
	pipe.HSetNX(r.buildElementPrefix(), session, 0)
	pipe.HIncrBy(r.buildElementPrefix(), session, 1)
	_, err = pipe.Exec()
	if err != nil {
		log.Error(err)
		return errRedis
	}
	cancelChannel := make(chan bool, 1)
	go r.collectElement(element, cancelChannel)
	cancelChannelMap[session] = cancelChannel

	return nil
}

func (r *RedisQueue) Ack(e pq.QueueElement) error {
	session := e.GetSession()
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
	return nil
}
