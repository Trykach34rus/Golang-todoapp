package core_goredis_pool

import (
	"errors"

	core_redis_pool "github.com/Trykach34rus/Golang-todoapp/internal/core/repository/redis/pool"
	"github.com/redis/go-redis/v9"
)

type goredisStringCmd struct {
	*redis.StringCmd
}


func (c goredisStringCmd)Bytes() ([]byte,error)  {
	data, err := c.StringCmd.Bytes()
	if err != nil {
		return nil, mapError(err)
	}

	return data,nil
}

type goredisStatusCmd struct {
	*redis.StatusCmd
}

type goredisIntCmd struct {
	*redis.IntCmd
}


func mapError(err error) error {
	if errors.Is(err,redis.Nil){
		return core_redis_pool.NotFound
	}

	return err
}