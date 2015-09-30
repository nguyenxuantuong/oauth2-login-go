package controllers

import (
	"github.com/garyburd/redigo/redis"
	"github.com/revel/revel"
	"auth/app/utils"
	"time"
)

var (
	Rpl *redis.Pool //redis pool connections
	RCache utils.RedisCache //redis cache instance -- keep pool connection + some serialization, deserialization sugar methods
)

//this just contain redis connection
type RedisController struct {
	Rdb redis.Conn
}

//init connection with redis using redigo redis
func InitRedis() {
	host := revel.Config.StringDefault("cache.redis.host", "127.0.0.1");
	password := revel.Config.StringDefault("cache.redis.password", "");
	DB := revel.Config.IntDefault("cache.redis.DB", 0);

	//initialize is similar to revel redis cache but allow to select-db -- so cannot reused the old code
	Rpl = &redis.Pool{
		MaxIdle:     revel.Config.IntDefault("cache.redis.maxidle", 5),
		MaxActive:   revel.Config.IntDefault("cache.redis.maxactive", 0),
		IdleTimeout: time.Duration(revel.Config.IntDefault("cache.redis.idletimeout", 240)) * time.Second,
		Dial: func() (redis.Conn, error) {
			protocol := revel.Config.StringDefault("cache.redis.protocol", "tcp")
			toc := time.Millisecond * time.Duration(revel.Config.IntDefault("cache.redis.timeout.connect", 10000))
			tor := time.Millisecond * time.Duration(revel.Config.IntDefault("cache.redis.timeout.read", 5000))
			tow := time.Millisecond * time.Duration(revel.Config.IntDefault("cache.redis.timeout.write", 5000))

			c, err := redis.DialTimeout(protocol, host, toc, tor, tow)

			if err != nil {
				return nil, err
			}

			if len(password) > 0 {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			} else {
				// check with PING
				if _, err := c.Do("PING"); err != nil {
					c.Close()
					return nil, err
				}
			}

			//then select db
			if _, err := c.Do("SELECT", DB); err != nil {
				c.Close()
				return nil, err
			}

			return c, err
		},
		// custom connection test method
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if _, err := c.Do("PING"); err != nil {
				return err
			}
			return nil
		},
	}

	//initialize an instance of revel redis cache -- which has some sugar methods like seriablization + deserialization
	defaultExpiration := time.Hour // The default for the default is one hour.
	if expireStr, found := revel.Config.String("cache.expires"); found {
		var err error
		if defaultExpiration, err = time.ParseDuration(expireStr); err != nil {
			panic("Could not parse default cache expiration duration " + expireStr + ": " + err.Error())
		}
	}

	//initialize an instance of redis revel cache
	RCache = utils.RedisCache{Rpl, defaultExpiration}

	//test if everything is fine
	conn := RCache.Pool.Get()
	if _, err := conn.Do("PING"); err != nil {
		panic("Unable to ping redis")
		revel.ERROR.Fatal("unable to ping redis")
	}
}

//those intercept function assign connection for each request
func (c *RedisController) Begin() revel.Result {
	c.Rdb = Rpl.Get()
	return nil
}

func (c *RedisController) End() revel.Result {
	//close the connections -- if it's initialize
	if c.Rdb != nil{
		c.Rdb.Close()
		c.Rdb = nil
	}
	
	return nil
}
