package main

import (
	"log"
	"os"
	"time"

	"github.com/garyburd/redigo/redis"
)

func main() {
	r, err := newRedis(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	if err := r.ping(); err != nil {
		log.Fatal(err)
	}
}

// Redis .
type Redis struct {
	clientPool *redis.Pool
}

func newRedis(host string) (*Redis, error) {
	// attempt to dial with given url to make sure we can reach redis
	client, err := redis.Dial("tcp", host)
	if err != nil {
		return nil, err
	}
	// connection was only open to test we can hit redis, close it
	client.Close()

	return &Redis{
		clientPool: &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 60 * time.Second,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", host)
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		},
	}, nil
}

func (r *Redis) ping() error {
	_, err := r.clientPool.Get().Do("PING")
	return err
}
