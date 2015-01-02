package controllers

import (
	"github.com/revel/revel"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"fmt"
)


// global variables -- such as to keep global db connections
var (
	Session *mgo.Session // Global mgo Session
	mgoSessionDupl func() *mgo.Session
	Dial    string       // http://godoc.org/labix.org/v2/mgo#Dial
	Method  string       // clone, copy, new http://godoc.org/labix.org/v2/mgo#Session.New
)

//this just contain mgo connections
type MgoController struct {
	//contain mongo session
	MongoSession *mgo.Session
}

func InitMgo(){
	var err error

	// Read configuration.
	Dial = revel.Config.StringDefault("revmgo.dial", "localhost")
	Method = revel.Config.StringDefault("revmgo.method", "clone")

	//make sure only some method are supported
	if Method != "clone" && Method != "copy" && Method != "new" {
		revel.ERROR.Fatal(fmt.Errorf("revmgo: Invalid session instantiation method '%s'", Method))
	}

	// Let's try to connect to Mongo DB right upon starting revel but don't
	// raise an error. Errors will be handled if there is actually a request
	if Session == nil {
		revel.INFO.Println("dial mongo")
		Session, err = mgo.Dial(Dial)

		if err != nil {
			// Only warn since we'll retry later for each request
			revel.ERROR.Fatal("Could not connect to Mongo DB. Error: %s", err)
		} else {
			switch Method {
			case "clone":
				mgoSessionDupl = Session.Clone
			case "copy":
				mgoSessionDupl = Session.Copy
			case "new":
				mgoSessionDupl = Session.New
			default:
				mgoSessionDupl = Session.Clone
			}
		}
	}

	// register the custom bson.ObjectId binder
	objId := bson.NewObjectId()
	revel.TypeBinders[reflect.TypeOf(objId)] = ObjectIdBinder
}


// Custom TypeBinder for bson.ObjectId
// Makes additional Id parameters in actions obsolete
var ObjectIdBinder = revel.Binder{
	// Make a ObjectId from a request containing it in string format.
	Bind: revel.ValueBinder(func(val string, typ reflect.Type) reflect.Value {
		if len(val) == 0 {
			return reflect.Zero(typ)
		}
		if bson.IsObjectIdHex(val) {
			objId := bson.ObjectIdHex(val)
			return reflect.ValueOf(objId)
		} else {
			revel.ERROR.Print("ObjectIdBinder.Bind - invalid ObjectId!")
			return reflect.Zero(typ)
		}
	}),
	// Turns ObjectId back to hexString for reverse routing
	Unbind: func(output map[string]string, name string, val interface{}) {
		var hexStr string
		hexStr = fmt.Sprintf("%s", val.(bson.ObjectId).Hex())
		// not sure if this is too carefull but i wouldn't want invalid ObjectIds in my App
		if bson.IsObjectIdHex(hexStr) {
			output[name] = hexStr
		} else {
			revel.ERROR.Print("ObjectIdBinder.Unbind - invalid ObjectId!")
			output[name] = ""
		}
	},
}

//those intercept function assign connection for each request
func (c *MgoController) Begin() revel.Result {
	//assign mongo session
	if Session == nil {
		var err error
		Session, err = mgo.Dial(Dial)

		if err != nil {
			// Extend the error description to include that this is a Mongo Error
			err = fmt.Errorf("Could not connect to Mongo DB. Error: %s", err)
			revel.ERROR.Printf("unable to connect to mongo %s", err)
			panic(err)
		} else {
			switch Method {
			case "clone":
				mgoSessionDupl = Session.Clone
			case "copy":
				mgoSessionDupl = Session.Copy
			case "new":
				mgoSessionDupl = Session.New
			default:
				mgoSessionDupl = Session.Clone
			}
		}
	}
	// Calls Clone(), Copy() or New() depending on the configuration
	c.MongoSession = mgoSessionDupl()

	return nil
}

func (c *MgoController) End() revel.Result {
	if c.MongoSession != nil {
		c.MongoSession.Close()
	}

	return nil
}
