package controllers
import (
	"github.com/revel/revel"
)

func init() {
	//initialize function -- which will initialize Mongo, Redis, MySQL
	revel.OnAppStart(Init) // invoke InitDB function before
	
	//intercept method before, after each request
	revel.InterceptMethod((*GormController).Begin, revel.BEFORE)
	revel.InterceptMethod((*GormController).Commit, revel.AFTER)
	revel.InterceptMethod((*GormController).Rollback, revel.FINALLY)
	
	//also intercept method for redis
	revel.InterceptMethod((*RedisController).Begin, revel.BEFORE)
	revel.InterceptMethod((*RedisController).End, revel.FINALLY)
	
	//and mgo
	revel.InterceptMethod((*MgoController).Begin, revel.BEFORE)
	revel.InterceptMethod((*MgoController).End, revel.FINALLY)
}
