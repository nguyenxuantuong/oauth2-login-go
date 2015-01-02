//gorm controllers
package controllers

import (
	"github.com/revel/revel"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"github.com/jinzhu/gorm"
	"auth/app/models"
	"fmt"
	"strings"
)

//this is global connections -- TODO: consider to move it into seperated packages?
var (
	Gdb gorm.DB // it can be used for jobs -- this is export field
)

//this just contain redis connection
type GormController struct {
	//this if for transaction type
	Txn *gorm.DB
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
		revel.ERROR.Fatal("FATAL", err)
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

// NOTE: Gorm controller will create a transaction
// This method fills the c.Txn before each transaction
func (c *GormController) Begin() revel.Result {
	txn := Gdb.Begin()
	if txn.Error != nil {
		panic(txn.Error)
	}
	c.Txn = txn
	return nil
}

// This method clears the c.Txn after each transaction
func (c *GormController) Commit() revel.Result {
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
func (c *GormController) Rollback() revel.Result {
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
