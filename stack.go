//Package hastack aims to share a thread-safe FIFO Stack across multiple hosts.
//HA Stack use Redis as backend and helps you to push and pop elements in your stack.
//See https://github.com/fsamin/go-hastack for sample usages.
package hastack

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

//Stack serves as a collection of elements, with two principal operations:
//push, which adds an element to the collection, and pop, which removes the most
//recently added element that was not yet removed. Hastack uses Redis ad backend
//and ensures thread-safe access to enable high availability capacities
//Import this as dependencies to manage a thread-safe FIFO Stack across
//multiple hosts
type Stack struct {
	name string
	pool *redis.Pool
}

//Connect to a stack and open its connection pool to redis
func Connect(name, redisServer, password string, maxIdle, idleTimeout int) (*Stack, error) {
	s := &Stack{name: name}
	s.pool = &redis.Pool{
		MaxIdle:     maxIdle,
		IdleTimeout: time.Duration(idleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisServer)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, e := c.Do("AUTH", password); err != e {
					c.Close()
					return nil, e
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	_, err := s.pool.Get().Do("PING")
	return s, err
}

//Push to the stack
func (s *Stack) Push(data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	c := s.pool.Get()
	defer c.Close()

	c.Send("MULTI")
	c.Send("INCR", s.name+":stats:counter")
	c.Send("LPUSH", s.name+":inbox", string(b))
	_, err = c.Do("EXEC")
	return err
}

//Pop from the stack
func (s *Stack) Pop(data interface{}) error {
	c := s.pool.Get()
	defer c.Close()
	res, err := c.Do("RPOPLPUSH", s.name+":inbox", s.name+":outbox")
	if err != nil {
		return err
	}
	if res != nil {
		if err := json.Unmarshal(res.([]byte), &data); err != nil {
			return err
		}
	}
	return nil
}

//BPop from the stack, this is block until there is something to pop
func (s *Stack) BPop(data interface{}) error {
	c := s.pool.Get()
	defer c.Close()
	var chanRes = make(chan string)
	var chanErr = make(chan error)

	go func() {
		for {
			res, err := c.Do("RPOPLPUSH", s.name+":inbox", s.name+":outbox")
			if err != nil {
				chanErr <- err
				break
			}
			if res != nil {
				s, err := redis.String(res, nil)
				if err != nil {
					chanErr <- err
					break
				}
				chanRes <- s
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case res := <-chanRes:
		if err := json.Unmarshal([]byte(res), data); err != nil {
			return err
		}
		return nil
	case err := <-chanErr:
		fmt.Println(err)
		return err
	}
}

//InboxLength returns size of pending message
func (s *Stack) InboxLength() (int64, error) {
	return s.length(s.name + ":inbox")
}

//OutboxLength returns size of computed message
func (s *Stack) OutboxLength() (int64, error) {
	return s.length(s.name + ":outbox")
}

func (s *Stack) length(str string) (int64, error) {
	c := s.pool.Get()
	res, err := c.Do("LLEN", str)
	if err != nil {
		return -1, err
	}
	return res.(int64), nil
}
