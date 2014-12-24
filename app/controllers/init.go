package controllers
import (
	"github.com/revel/revel"
)

func init() {
	revel.OnAppStart(InitDB) // invoke InitDB function before

	revel.InterceptMethod((*GormController).Begin, revel.BEFORE)
	revel.InterceptMethod((*GormController).Commit, revel.AFTER)
	revel.InterceptMethod((*GormController).Rollback, revel.FINALLY)
}
