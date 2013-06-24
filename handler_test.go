// Copyright 2013 mongoapi authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	. "launchpad.net/gocheck"
	"net/http"
	"net/http/httptest"
	"os"
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

func (s *S) TestBindShouldReturnsLocalhostWhenThePublicHostEnvIsNil(c *C) {
	request, err := http.NewRequest("POST", "/resources/myapp?:name=myapp", nil)
	os.Setenv("MONGODB_PUBLIC_URI", "")
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	err = Bind(recorder, request)
	c.Assert(err, IsNil)
	defer func() {
		database := session().DB("myapp")
		database.RemoveUser("myapp")
		database.DropDatabase()
	}()
	c.Assert(recorder.Code, Equals, http.StatusCreated)
	result, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, IsNil)
	data := map[string]string{}
	json.Unmarshal(result, &data)
	c.Assert(data["MONGO_URI"], Equals, "127.0.0.1:27017")
	c.Assert(data["MONGO_USER"], Equals, "myapp")
	c.Assert(data["MONGO_DATABASE_NAME"], Equals, "myapp")
	c.Assert(data["MONGO_PASSWORD"], Not(HasLen), 0)
}

func (s *S) TestBindWithReplicaSet(c *C) {
	request, err := http.NewRequest("POST", "/resources/myapp?:name=myapp", nil)
	publicHost := "mongoapi.com:27017"
	os.Setenv("MONGODB_PUBLIC_URI", publicHost)
	os.Setenv("MONGODB_REPLICA_SET", "tsuru")
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	err = Bind(recorder, request)
	c.Assert(err, IsNil)
	defer func() {
		database := session().DB("myapp")
		database.RemoveUser("myapp")
		database.DropDatabase()
	}()
	c.Assert(recorder.Code, Equals, http.StatusCreated)
	var data map[string]string
	err = json.NewDecoder(recorder.Body).Decode(&data)
	c.Assert(err, IsNil)
	c.Assert(data["MONGO_REPLICA_SET"], Equals, "tsuru")
}

func (s *S) TestBindShouldReturnTheVariables(c *C) {
	request, err := http.NewRequest("POST", "/resources/myapp?:name=myapp", nil)
	publicHost := "mongoapi.com:27017"
	os.Setenv("MONGODB_PUBLIC_URI", publicHost)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	err = Bind(recorder, request)
	c.Assert(err, IsNil)
	defer func() {
		database := session().DB("myapp")
		database.RemoveUser("myapp")
		database.DropDatabase()
	}()
	c.Assert(recorder.Code, Equals, http.StatusCreated)
	result, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, IsNil)
	data := map[string]string{}
	json.Unmarshal(result, &data)
	c.Assert(data["MONGO_URI"], Equals, publicHost)
	c.Assert(data["MONGO_USER"], Equals, "myapp")
	c.Assert(data["MONGO_DATABASE_NAME"], Equals, "myapp")
	c.Assert(data["MONGO_PASSWORD"], Not(HasLen), 0)
}

func (s *S) TestBindShouldCreateTheDatabase(c *C) {
	request, err := http.NewRequest("POST", "/resources/myapp?:name=myapp", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	err = Bind(recorder, request)
	c.Assert(err, IsNil)
	defer func() {
		database := session().DB("myapp")
		database.RemoveUser("myapp")
		database.DropDatabase()
	}()
	c.Assert(recorder.Code, Equals, http.StatusCreated)
	databases, err := session().DatabaseNames()
	c.Assert("myapp", In, databases)
}

func (s *S) TestBindShouldAddUserInTheDatabase(c *C) {
	request, err := http.NewRequest("POST", "/resources/myapp?:name=myapp", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	err = Bind(recorder, request)
	c.Assert(err, IsNil)
	defer func() {
		database := session().DB("myapp")
		database.RemoveUser("myapp")
		database.DropDatabase()
	}()
	c.Assert(recorder.Code, Equals, http.StatusCreated)
	collection := session().DB("myapp").C("system.users")
	lenght, err := collection.Find(bson.M{"user": "myapp"}).Count()
	c.Assert(lenght, Equals, 1)
}

func (s *S) TestUnbindShouldRemoveTheUser(c *C) {
	name := "myapp"
	database := session().DB(name)
	database.AddUser(name, "", false)
	defer func() {
		database.DropDatabase()
	}()
	request, err := http.NewRequest("DELETE", "/resources/myapp/hostname/10.10.10.10?:name=myapp&hostname=10.10.10.10", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	err = Unbind(recorder, request)
	c.Assert(err, IsNil)
	c.Assert(recorder.Code, Equals, http.StatusOK)
	collection := session().DB(name).C("system.users")
	lenght, err := collection.Find(bson.M{"user": name}).Count()
	c.Assert(lenght, Equals, 0)
}

func (s *S) TestRemoveShouldRemovesTheDatabase(c *C) {
	name := "myapp"
	database := session().DB(name)
	database.AddUser(name, "", false)
	request, err := http.NewRequest("DELETE", "/resources/name?:name=myapp", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	err = Remove(recorder, request)
	c.Assert(err, IsNil)
	c.Assert(recorder.Code, Equals, http.StatusOK)
	databases, err := session().DatabaseNames()
	c.Assert(name, Not(In), databases)
}

func (s *S) TestStatus(c *C) {
	name := "myapp"
	database := session().DB(name)
	database.AddUser(name, "", false)
	defer func() {
		database.RemoveUser("myapp")
		database.DropDatabase()
	}()
	request, err := http.NewRequest("GET", "/resources/myapp/status?:name=myapp", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	err = Status(recorder, request)
	c.Assert(err, IsNil)
	c.Assert(recorder.Code, Equals, http.StatusNoContent)
}

func errorHandler(w http.ResponseWriter, r *http.Request) error {
	return stderrors.New("some error")
}

func simpleHandler(w http.ResponseWriter, r *http.Request) error {
	fmt.Fprint(w, "success")
	return nil
}

func (s *S) TestHandlerReturns500WhenInternalHandlerReturnsAnError(c *C) {
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/apps", nil)
	c.Assert(err, IsNil)

	Handler(errorHandler).ServeHTTP(recorder, request)
	c.Assert(recorder.Code, Equals, 500)
	c.Assert(recorder.Body.String(), Equals, "some error\n")
}

func (s *S) TestHandlerShouldPassAnHandlerWithoutError(c *C) {
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/apps", nil)
	c.Assert(err, IsNil)

	Handler(simpleHandler).ServeHTTP(recorder, request)
	c.Assert(recorder.Code, Equals, 200)
	c.Assert(recorder.Body.String(), Equals, "success")
}
