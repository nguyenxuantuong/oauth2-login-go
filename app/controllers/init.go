package controllers
import (
	"github.com/revel/revel"
)

func init() {
	//initialize function
	revel.OnAppStart(Init) // invoke InitDB function before
	
	//intercept method before, after each request
	revel.InterceptMethod((*GormController).Begin, revel.BEFORE)
	revel.InterceptMethod((*GormController).Commit, revel.AFTER)
	revel.InterceptMethod((*GormController).Rollback, revel.FINALLY)
}
