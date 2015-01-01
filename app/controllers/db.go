package controllers

import (
	"github.com/jinzhu/gorm"
	"github.com/garyburd/redigo/redis"
	r "github.com/revel/revel"
//	"github.com/revel/revel/cache"
	_ "github.com/go-sql-driver/mysql"
	"github.com/revel/revel"
	"auth/app/models"
	"database/sql"
	"strings"
	"fmt"
	"time"
	"auth/app/utils"
)

// type: revel controller with `*gorm.DB`
// c.Txn will keep `Gdb *gorm.DB`
type GormController struct {
	*r.Controller
	//this if for transaction type
	Txn *gorm.DB
	//this is contain an connection
	Rdb redis.Conn
}

//interface for API response
type Response struct {
	Status string `json:"status"`
	Data   interface{} `json:"data"`
	Errors	interface {} `json:"errors"`
}

//sugar function to be shared among other controllers
func (c GormController) RenderJsonError(errors interface {}) r.Result{
	return c.RenderJson(Response{Status: "error", Errors: errors})
}

//sugar function to return success data
func (c GormController) RenderJsonSuccess(data interface {}) r.Result {
	return c.RenderJson(Response{Status: "success", Data: data})
}

// it can be used for jobs -- this is export field
var Gdb gorm.DB
var Rpl *redis.Pool
var RCache utils.RedisCache
var SessionExpire time.Duration

func Init(){
	//init session expire variable
	var expireAfterDuration time.Duration

	//keep the expire duration key
	var err error
	if expiresString, ok := r.Config.String("session.expires"); !ok {
		expireAfterDuration = 30 * 24 * time.Hour
	} else if expiresString == "session" {
		expireAfterDuration = 0
	} else if expireAfterDuration, err = time.ParseDuration(expiresString); err != nil {
		panic(fmt.Errorf("session.expires invalid: %s", err))
	}

	SessionExpire = expireAfterDuration
	
	//init db connections
	InitDB()
	InitRedis()
}

//init MySQL connection
func InitDB() {
	host := revel.Config.StringDefault("db.host", "127.0.0.1")
	port := revel.Config.StringDefault("db.port", "3306")
	user := revel.Config.StringDefault("db.user", "")
	pass := revel.Config.StringDefault("db.password", "")
	dbname := revel.Config.StringDefault("db.name", "go-auth")
	protocol := revel.Config.StringDefault("db.protocol", "tcp")

	//NOTE: charset=utf8 is needed; otherwise, it will have datatime read/write problem from utf8 to Time.time
	dbargs := revel.Config.StringDefault("dbargs", "charset=utf8&parseTime=true")

	if strings.Trim(dbargs, " ") != "" {
		dbargs = "?" + dbargs
	} else {
		dbargs = ""
	}
	
	connectionString := fmt.Sprintf("%s:%s@%s([%s]:%s)/%s%s",
		user, pass, protocol, host, port, dbname, dbargs)
	
	var err error
	Gdb, err = gorm.Open("mysql", connectionString)

	if err != nil {
		r.ERROR.Fatal("FATAL", err)
		panic( err )
	}

	//some settings
	Gdb.DB().SetMaxIdleConns(10)
	Gdb.DB().SetMaxOpenConns(100)

	// Disable table name's pluralization
	Gdb.SingularTable(true)
	
	//auto-migrate models
	Gdb.AutoMigrate(&models.User{})
	Gdb.AutoMigrate(&models.AccountActivation{})
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
		r.ERROR.Fatal("unable to ping redis")
	}
}


// TODO: check if it really create transaction; only use transaction if needed
// This method fills the c.Txn before each transaction
func (c *GormController) Begin() r.Result {
	//assign c.Rdb before each transaction
	defer func(){
		c.Rdb = Rpl.Get()
	}()
	
	txn := Gdb.Begin()
	if txn.Error != nil {
		panic(txn.Error)
	}
	c.Txn = txn
	return nil
}

// This method clears the c.Txn after each transaction
func (c *GormController) Commit() r.Result {
	//clear the pointer from controllers
	defer func(){
		//close the connections -- if it's initialize
		if c.Rdb != nil{
			c.Rdb.Close()
			c.Rdb = nil
		}
	}()
	
	if c.Txn == nil {
		return nil
	}
	c.Txn.Commit()
	if err := c.Txn.Error; err != nil && err != sql.ErrTxDone {
		panic(err)
	}
	c.Txn = nil
	return nil
}

// This method clears the c.Txn after each transaction, too
func (c *GormController) Rollback() r.Result {
	//clear pointer from the controller
	defer func(){
		if c.Rdb != nil{
			c.Rdb.Close()
			c.Rdb = nil
		}
	}()
	
	if c.Txn == nil {
		return nil
	}
	c.Txn.Rollback()
	if err := c.Txn.Error; err != nil && err != sql.ErrTxDone {
		panic(err)
	}
	c.Txn = nil
	return nil
}
