package main

import (
	"encoding/json"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	. "launchpad.net/gocheck"
	"net/http"
	"net/http/httptest"
	"testing"
)

var _ = Suite(&S{})

type S struct{}

func Test(t *testing.T) { TestingT(t) }

type InChecker struct{}

func (c *InChecker) Info() *CheckerInfo {
	return &CheckerInfo{Name: "In", Params: []string{"value", "list"}}
}

func (c *InChecker) Check(params []interface{}, names []string) (bool, string) {
	if len(params) != 2 {
		return false, "you should provide two parameters"
	}
	value, ok := params[0].(string)
	if !ok {
		return false, "first parameter should be a string"
	}
	list, ok := params[1].([]string)
	if !ok {
		return false, "second parameter should be a slice"
	}
	for _, item := range list {
		if value == item {
			return true, ""
		}
	}
	return false, ""
}

var In Checker = &InChecker{}

func (s *S) TestAdd(c *C) {
	request, err := http.NewRequest("POST", "/resources/", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	Add(recorder, request)
	c.Assert(recorder.Code, Equals, http.StatusCreated)
}

func (s *S) TestBindShouldReturnsTheVariables(c *C) {
	request, err := http.NewRequest("POST", "/resources/myapp?:name=myapp", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	Bind(recorder, request)
	defer func() {
		database := Session.DB("myapp")
		database.RemoveUser("myapp")
		database.DropDatabase()
	}()
	c.Assert(recorder.Code, Equals, http.StatusCreated)
	result, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, IsNil)
	expected := map[string]string{
		"MONGO_URI":           "127.0.0.1:27017",
		"MONGO_USER":          "myapp",
		"MONGO_PASSWORD":      "",
		"MONGO_DATABASE_NAME": "myapp",
	}
	data := map[string]string{}
	json.Unmarshal(result, &data)
	c.Assert(data, DeepEquals, expected)
}

func (s *S) TestBindShouldCreateTheDatabase(c *C) {
	request, err := http.NewRequest("POST", "/resources/myapp?:name=myapp", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	Bind(recorder, request)
	defer func() {
		database := Session.DB("myapp")
		database.RemoveUser("myapp")
		database.DropDatabase()
	}()
	c.Assert(recorder.Code, Equals, http.StatusCreated)
	databases, err := Session.DatabaseNames()
	c.Assert("myapp", In, databases)
}

func (s *S) TestBindShouldAddUserInTheDatabase(c *C) {
	request, err := http.NewRequest("POST", "/resources/myapp?:name=myapp", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	Bind(recorder, request)
	defer func() {
		database := Session.DB("myapp")
		database.RemoveUser("myapp")
		database.DropDatabase()
	}()
	c.Assert(recorder.Code, Equals, http.StatusCreated)
	collection := Session.DB("myapp").C("system.users")
	lenght, err := collection.Find(bson.M{"user": "myapp"}).Count()
	c.Assert(lenght, Equals, 1)
}

func (s *S) TestUnbindShouldRemoveTheUser(c *C) {
	name := "myapp"
	database := Session.DB(name)
	database.AddUser(name, "", false)
	defer func() {
		database.DropDatabase()
	}()
	request, err := http.NewRequest("DELETE", "/resources/myapp/hostname/10.10.10.10?:name=myapp&hostname=10.10.10.10", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	Unbind(recorder, request)
	c.Assert(recorder.Code, Equals, http.StatusOK)
	collection := Session.DB(name).C("system.users")
	lenght, err := collection.Find(bson.M{"user": name}).Count()
	c.Assert(lenght, Equals, 0)
}

func (s *S) TestRemoveShouldRemovesTheDatabase(c *C) {
	name := "myapp"
	database := Session.DB(name)
	database.AddUser(name, "", false)
	request, err := http.NewRequest("DELETE", "/resources/name?:name=myapp", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	Remove(recorder, request)
	c.Assert(recorder.Code, Equals, http.StatusOK)
	databases, err := Session.DatabaseNames()
	c.Assert(name, Not(In), databases)
}

func (s *S) TestStatus(c *C) {
	name := "myapp"
	database := Session.DB(name)
	database.AddUser(name, "", false)
	defer func() {
		database.RemoveUser("myapp")
		database.DropDatabase()
	}()
	request, err := http.NewRequest("GET", "/resources/myapp/status?:name=myapp", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	Status(recorder, request)
	c.Assert(recorder.Code, Equals, http.StatusNoContent)
}

func (s *S) TestStatusShouldReturns500WhenMongoIsNotUp(c *C) {
	request, err := http.NewRequest("GET", "/resources/myapp/status?:name=myapp", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	Status(recorder, request)
	c.Assert(recorder.Code, Equals, http.StatusInternalServerError)
}
