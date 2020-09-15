package cache

import (
	"net/url"
	"strconv"
	"sync"

	"github.com/go-redis/redis"
	"github.com/golang/glog"

	"github.com/gisvr/golib/log"
)

var (
	redisOnce    sync.Once
	redisClients map[string]*redis.Client
)

const (
	__defaultName     = "default"
	__defaultHostPort = "127.0.0.1:6379"
)

// redis://db:password@host:port
func Init(addr string) {
	redisOnce.Do(func() {
		glog.V(8).Infof("redis client init. addrs => %s", addr)
		redisClients = make(map[string]*redis.Client)
		createOneRedisClient(__defaultName, addr)
	})
}

func Get() *redis.Client {
	Init(__defaultHostPort)
	return redisClients[__defaultName]
}

// redis://db:password@host:port
func parseRedisAddr(addr string) (host string, password string, db int) {
	db = 0
	u, err := url.Parse(addr)
	if err != nil {
		host = addr
	} else {
		host = u.Host
		db64, _ := strconv.ParseInt(u.User.Username(), 0, 32)

		db = int(db64)
		password, _ = u.User.Password()
	}

	log.Infof("parse redis URI. addr = %s, host = %s, password = %s, db = %d", addr, host, password, db)
	return
}

func createOneRedisClient(name, addr string) {
	host, password, db := parseRedisAddr(addr)
	redisClients[name] = redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password, // no password set
		DB:       db,       // use default DB
	})
}
