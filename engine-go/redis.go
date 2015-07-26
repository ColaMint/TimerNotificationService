package main

import (
	"fmt"
	"github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
	"time"
)

var RedisPool *redis.Pool

func InitRedisPool() {
	RedisPool = &redis.Pool{
		MaxIdle:     4,
		IdleTimeout: 60 * time.Second,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")

			return err
		},
		Dial: func() (redis.Conn, error) {
			redisConfig := &Config.RedisConfig
			c, err := redis.Dial("tcp", fmt.Sprintf("%v:%v", redisConfig.RedisHost, redisConfig.RedisPort))
			if err != nil {
				seelog.Errorf("[Redis Connection Error] %v", err)
				return nil, err
			}
			_, err = c.Do("SELECT", redisConfig.RedisDB)
			return c, err
		},
	}
}
