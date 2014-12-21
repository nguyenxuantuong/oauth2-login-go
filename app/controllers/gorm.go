package controllers

import (
	"github.com/jinzhu/gorm"
	r "github.com/revel/revel"
	_ "github.com/go-sql-driver/mysql"
	"auth/app/models"
	"database/sql"
	"strings"
	"fmt"
)

// type: revel controller with `*gorm.DB`
// c.Txn will keep `Gdb *gorm.DB`
type GormController struct {
	*r.Controller
	Txn *gorm.DB
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

func getParamString(param string, defaultValue string) string {
	p, found := r.Config.String(param)
	if !found {
		if defaultValue == "" {
			r.ERROR.Fatal("Cound not find parameter: " + param)
		} else {
			return defaultValue
		}
	}
	return p
}

func getConnectionString() string {
	host := getParamString("db.host", "")
	port := getParamString("db.port", "3306")
	user := getParamString("db.user", "")
	pass := getParamString("db.password", "")
	dbname := getParamString("db.name", "auction")
	protocol := getParamString("db.protocol", "tcp")
	
	//NOTE: charset=utf8 is needed; otherwise, it will have datatime read/write problem from utf8 to Time.time
	dbargs := getParamString("dbargs", "charset=utf8&parseTime=true")

	if strings.Trim(dbargs, " ") != "" {
		dbargs = "?" + dbargs
	} else {
		dbargs = ""
	}
	return fmt.Sprintf("%s:%s@%s([%s]:%s)/%s%s",
		user, pass, protocol, host, port, dbname, dbargs)
}

func InitDB() {
	connectionString := getConnectionString()
	
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


// TODO: check if it really create transaction; only use transaction if needed
// This method fills the c.Txn before each transaction
func (c *GormController) Begin() r.Result {
	txn := Gdb.Begin()
	if txn.Error != nil {
		panic(txn.Error)
	}
	c.Txn = txn
	return nil
}

// This method clears the c.Txn after each transaction
func (c *GormController) Commit() r.Result {
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
