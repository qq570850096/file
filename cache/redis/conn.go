package redis

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)

var (
	pool      *redis.Pool
	redisHost = "127.0.0.1:6379"
	redisPass = "testupload"
)

// newRedisPool:创建连接池
func newRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,
		MaxActive:   30,
		IdleTimeout: 30 * time.Second,
		Dial: func() (conn redis.Conn, e error) {
			// 1. 打开连接
			conn, e = redis.Dial("tcp", redisHost)
			if e != nil {
				fmt.Println(e)
				return nil, e
			}
			// 2. 访问认证
			if _, e = conn.Do("AUTH", redisPass); e != nil {
				conn.Close()
				fmt.Println(e)
				return nil, e
			}
			return conn, e
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			// 每分钟检测一次可用性
			if time.Since(t) < time.Minute {
				return nil
			}
			// 超过一分钟用 Ping 检测一次
			_, err := c.Do("PING")
			return err
		},
	}
}

func init() {
	pool = newRedisPool()
}
// 对外提供服务的接口
func RedisPool() *redis.Pool {
	return pool
}
