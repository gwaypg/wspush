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
	dbIndex := etc.Etc.String("module/cache/redis", "dbi")
	password := etc.Etc.String("module/cache/redis", "passwd")
	poolSize := etc.Etc.Int64("module/cache/redis", "pool_size")
	if len(dbIndex) == 0 {
		dbIndex = "0"
	}
	r, err := redis.NewRediStoreWithDB(int(poolSize), "tcp", host, password, dbIndex)
	if err != nil {
		panic(errors.As(err, host, password, poolSize, dbIndex))
	}
	gRedis = r
}

func GetRedis() *redis.RediStore {
	if gRedis == nil {
		panic("redis not init")
	}
	return gRedis
}
