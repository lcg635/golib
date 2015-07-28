package storage

import (
	"encoding/json"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

type RedisCommondRunner func() error

type RedisPool struct {
	pool        *redis.Pool
	server      string
	password    string
	maxIdle     int
	maxActive   int
	idleTimeout time.Duration
}

func NewDefaultRedisPool(server, password string) *RedisPool {
	return NewRedisPool(server, password, 20, 20, 240*time.Second)
}

func NewRedisPool(server, password string, maxActive, maxIdle int, idleTimeout time.Duration) *RedisPool {
	return &RedisPool{
		server:      server,
		password:    password,
		maxIdle:     maxIdle,
		maxActive:   maxActive,
		idleTimeout: idleTimeout,
	}
}

func (this *RedisPool) RedisPool() *redis.Pool {
	if this.pool == nil {
		this.pool = &redis.Pool{
			MaxIdle:     this.maxIdle,
			IdleTimeout: this.idleTimeout,
			MaxActive:   this.maxActive,
			Wait:        true,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", this.server)
				if err != nil {
					return nil, err
				}
				if this.password != "" {
					if _, err := c.Do("AUTH", this.password); err != nil {
						c.Close()
						return nil, err
					}
				}
				return c, nil
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		}
	}
	return this.pool
}

/**
 * 缓存一定时间
 */
func (this *RedisPool) GetCachedEx(res interface{}, key string, expire int, f RedisCommondRunner) error {
	c := this.RedisPool().Get()
	// defer c.Close()

	reply, err := redis.Bytes(c.Do("GET", key))
	if err == nil && reply != nil {
		c.Close()
		return json.Unmarshal(reply, res)
	}

	if err = f(); err != nil {
		c.Close()
		return err
	}

	var buf []byte
	if buf, err = json.Marshal(res); err != nil {
		c.Close()
		return err
	}

	_, err = c.Do("SETEX", key, expire, buf)
	if err != nil {
		log.Println("REDIS SETEX", key, "failed", err)
	}

	c.Close()
	return nil
}

///**
// * 定长列表添加，可用于排行、历史记录等
// */
//func AddToLimitedListWithSortedSet(key string, limit int64, f func() (int64, interface{}, error)) error {
//	var score int64, value interface{}, err error
//	if score, value, err = f(); err != nil {
//		return err
//	}

//	return this.Do(nil, func(c redis.Conn) error {
//		_, err = c.Do("ZADD", key, score, value)
//		if err != nil {
//			return err
//		}

//		length, err := redis.Int64(c.Do("ZCARD", key))
//		if err != nil {
//			log.Println("REDIS ZCARD", key, "failed", err)
//		}

//		if length >= limit {
//			_, err := c.Do("ZREMRANGEBYRANK", key, 0, 0)
//			if err != nil {
//				log.Println("REDIS ZREMRANGEBYRANK", key, "failed", err)
//			}
//		}

//		return nil
//	})
//}

//func GetLimitedListWithSortedSet(key string, take int, desc bool) ([]string, error) {
//	stop := -1
//	if take > 0 {
//		stop = take + 1
//	}

//	opt := "ZREVRANGE"
//	if desc != true {
//		opt = "ZRANGE"
//	}

//	var (
//		reply []string
//		err   error
//	)
//	err = RedisDo(func(c redis.Conn) error {
//		reply, err = redis.Strings(c.Do(opt, key, 0, stop))
//		return err
//	})

//	return reply, err
//}
