package cache

import (
	"github.com/gwaylib/errors"
	"github.com/gwaylib/redis"
	"github.com/gwaypg/wspush/module/etc"
)

var gRedis *redis.RediStore

const (
	REDIS_WSNODE_CALLBACK_PREFIX = "_wsnode_callback_"
)

func init() {
	host := etc.Etc.String("module/cache/redis", "uri")
	password := etc.Etc.String("module/cache/redis", "passwd")
	poolSize := etc.Etc.Int64("module/cache/redis", "pool_size")
	r, err := redis.NewRediStore(int(poolSize), "tcp", host, password)
	if err != nil {
		panic(errors.As(err, host, password, poolSize))
	}
	gRedis = r
}

func GetRedis() *redis.RediStore {
	if gRedis == nil {
		panic("redis not init")
	}
	return gRedis
}
