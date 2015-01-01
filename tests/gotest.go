//NOTE: all of the test method has to started with Test prefix; otherwise, it will not being run
package tests

import (
	"github.com/revel/revel"
	"fmt"
	"encoding/json"
	"bytes"
	"strings"
)

var _ = fmt.Printf
var _  = bytes.Index
var _ = strings.Index

type GoTest struct {
	revel.TestSuite
}

//multiple way to declare struct
func (t *GoTest) TestStruct() {
	type Student struct{
		Name string
		Age int
	}

	//initialize struct value
	var student = Student{Name: "Nguyen", Age: 12}
	t.AssertEqual(student.Name, "Nguyen")
	t.AssertEqual(student.Age, 12)

	//using pointer
	var studentPtr  *Student
	studentPtr = &Student{Name: "Nguyen", Age: 12}
	t.AssertEqual(studentPtr.Name, "Nguyen")

	var anonymousStudent struct {
		Name 	string
		Age 	int
	}
	anonymousStudent.Name = "Nguyen"
	t.AssertEqual(anonymousStudent.Name, "Nguyen")

	var student2 = struct {
			Name string
		 	Age int
	}{
		"Nguyen",
		12,
	}

	t.AssertEqual(student2.Name, "Nguyen")
}

//unmarshal json
func (t *GoTest) TestDecodeJsonSimple(){
	type Item struct {
		Title string
		URL string
	}

	var itemData = `{"Title": "The Go homepage", "URL": "http://golang.org/"}`;
	item := Item{}
	json.Unmarshal([]byte(itemData), &item)
	t.AssertEqual(item.Title, "The Go homepage");
	t.AssertEqual(item.URL, "http://golang.org/");
}

//map custom key name for unmarshalling json
func (t *GoTest) TestDecodeJsonFieldName(){
	type Item struct {
		Title string `json:"titleCustom"`
		URL string  `json:"urlCustom"`
	}

	var itemData = `{"titleCustom": "The Go homepage", "urlCustom": "http://golang.org/"}`;
	item := Item{}
	json.Unmarshal([]byte(itemData), &item)
	t.AssertEqual(item.Title, "The Go homepage");
	t.AssertEqual(item.URL, "http://golang.org/");
}

//unmarshalling json accepting case insensitive
func (t *GoTest) TestDecodeJsonCaseInsensitive(){
	type Item struct {
		Title string
		URL string
	}

	var itemData = `{"title": "The Go homepage", "url": "http://golang.org/"}`;
	item := Item{}
	json.Unmarshal([]byte(itemData), &item)
	t.AssertEqual(item.Title, "The Go homepage");
	t.AssertEqual(item.URL, "http://golang.org/");
}

//trying to decode complex data
func (t *GoTest) TestDecodeJson(){
	type Item struct {
		Title string
		URL string
	}

	type Response struct {
		Data struct {
			Children []struct{
				Data Item
			}
		}
	}

	var jsonData = `{"Data": {"Children": [{"Data": {"Title": "The Go homepage", "URL": "http://golang.org/"}}]}}`;

	r := Response{}
	json.Unmarshal([]byte(jsonData), &r)

	t.AssertEqual(r.Data.Children[0].Data.Title, "The Go homepage");
	t.AssertEqual(r.Data.Children[0].Data.URL, "http://golang.org/");
	t.AssertEqual(len(r.Data.Children), 1);
}


//it also will be unmarshal to different keyname
func (t *GoTest) TestMarshallJsonDifferentKey(){
	type Item struct {
		Title string `json:"titleCustom"`
		URL string  `json:"urlCustom"`
	}

	item := Item{
		Title: "Hello",
		URL: "www.google.com",
	}

	b, err := json.Marshal(item)
	t.AssertEqual(err, nil)

	var s = string(b)
	t.AssertNotEqual(strings.Index(s, "titleCustom"), -1)
	t.AssertNotEqual(strings.Index(s, "urlCustom"), -1)
}

type Valueable interface {
	Value() int64
}

type Parent struct{
	value int64
}

func (i *Parent) Value() int64 {
	return i.value
}

type Child struct{
	Parent
	multiplier int64
}

func (i Child) Value() int64 {
	return i.value * i.multiplier
}

//NOTE: we need to use value overhere; not pointer?
func GetValue(c Valueable) int64{
	return c.Value();
}

func (t *GoTest) TestEmbedByValue(){
	var parent = Parent{3};
	var child = &Child{parent, 3}
	
	t.AssertEqual(parent.Value(), 3);
	t.AssertEqual(child.Value(), 9);
	t.AssertEqual(child.Parent.Value(), 3);
	
	//this using instance instead of value
	var childInstance = Child{parent, 3}
	t.AssertEqual(childInstance.Parent.Value(), 3);
	
	//parent using pointer; so need to use &parent; otherwise, interface won't work
	//t.AssertEqual(GetValue(parent), 3); --> this won't work because parent using pointer for overloading Value() of interface
	t.AssertEqual(GetValue(&parent), 3);

	//child method using value; not pointer; so need to put value here
	t.AssertEqual(GetValue(child), 9);
	t.AssertEqual(GetValue(childInstance), 9); //->this work because child using value; not pointer when overloading Value() of interface
}
