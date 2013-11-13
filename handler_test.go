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
	"launchpad.net/gocheck"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var _ = gocheck.Suite(&S{})

type S struct {
	muxer http.Handler
}

func Test(t *testing.T) { gocheck.TestingT(t) }

func (s *S) SetUpSuite(c *gocheck.C) {
	s.muxer = buildMux()
}

func (s *S) TearDownSuite(c *gocheck.C) {
	session().DB(dbName()).DropDatabase()
}

type InChecker struct{}

func (c *InChecker) Info() *gocheck.CheckerInfo {
	return &gocheck.CheckerInfo{Name: "In", Params: []string{"value", "list"}}
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

var In gocheck.Checker = &InChecker{}

func (s *S) TestAdd(c *gocheck.C) {
	body := strings.NewReader("name=something")
	request, err := http.NewRequest("POST", "/resources", body)
	c.Assert(err, gocheck.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, gocheck.Equals, http.StatusCreated)
}

func (s *S) TestAddReservedName(c *gocheck.C) {
	name := dbName()
	body := strings.NewReader("name=" + name)
	request, err := http.NewRequest("POST", "/resources", body)
	c.Assert(err, gocheck.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, gocheck.Equals, http.StatusForbidden)
	c.Assert(recorder.Body.String(), gocheck.Equals, "Reserved name")
}

func (s *S) TestBindShouldReturnLocalhostWhenThePublicHostEnvIsNil(c *gocheck.C) {
	body := strings.NewReader("app-host=localhost&unit-host=127.0.0.1")
	request, err := http.NewRequest("POST", "/resources/myapp", body)
	c.Assert(err, gocheck.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	os.Setenv("MONGODB_PUBLIC_URI", "")
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	defer func() {
		database := session().DB("myapp")
		database.RemoveUser("myapp")
		database.DropDatabase()
	}()
	c.Assert(recorder.Code, gocheck.Equals, http.StatusCreated)
	result, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, gocheck.IsNil)
	data := map[string]string{}
	json.Unmarshal(result, &data)
	c.Assert(data["MONGO_URI"], gocheck.Equals, "127.0.0.1:27017")
	c.Assert(data["MONGO_USER"], gocheck.Equals, "myapp")
	c.Assert(data["MONGO_DATABASE_NAME"], gocheck.Equals, "myapp")
	c.Assert(data["MONGO_PASSWORD"], gocheck.Not(gocheck.HasLen), 0)
	coll := collection()
	expected := dbBind{
		AppHost:  "localhost",
		Name:     "myapp",
		Password: data["MONGO_PASSWORD"],
		Units:    []string{"127.0.0.1"},
	}
	var bind dbBind
	q := bson.M{"name": "myapp"}
	defer coll.Remove(q)
	err = coll.Find(q).One(&bind)
	c.Assert(err, gocheck.IsNil)
	c.Assert(bind, gocheck.DeepEquals, expected)
}

func (s *S) TestBindTwoInstances(c *gocheck.C) {
	body := strings.NewReader("app-host=localhost&unit-host=127.0.0.1")
	request, err := http.NewRequest("POST", "/resources/myapp?:name=myapp", body)
	c.Assert(err, gocheck.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	os.Setenv("MONGODB_PUBLIC_URI", "")
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	defer func() {
		database := session().DB("myapp")
		database.RemoveUser("myapp")
		database.DropDatabase()
		collection().RemoveAll(bson.M{"name": "myapp"})
	}()
	var first, second map[string]string
	err = json.NewDecoder(recorder.Body).Decode(&first)
	c.Assert(err, gocheck.IsNil)
	body = strings.NewReader("app-host=localhost&unit-host=127.0.0.2")
	request, err = http.NewRequest("POST", "/resources/myapp", body)
	c.Assert(err, gocheck.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder = httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, gocheck.Equals, http.StatusCreated)
	err = json.NewDecoder(recorder.Body).Decode(&second)
	c.Assert(err, gocheck.IsNil)
	c.Assert(second, gocheck.DeepEquals, first)
	expected := dbBind{
		AppHost:  "localhost",
		Name:     "myapp",
		Password: first["MONGO_PASSWORD"],
		Units:    []string{"127.0.0.1", "127.0.0.2"},
	}
	var bind dbBind
	err = collection().Find(bson.M{"name": "myapp"}).One(&bind)
	c.Assert(err, gocheck.IsNil)
	c.Assert(bind, gocheck.DeepEquals, expected)
}

func (s *S) TestBindWithReplicaSet(c *gocheck.C) {
	body := strings.NewReader("app-host=localhost&unit-host=127.0.0.1")
	request, err := http.NewRequest("POST", "/resources/myapp", body)
	c.Assert(err, gocheck.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	publicHost := "mongoapi.com:27017"
	os.Setenv("MONGODB_PUBLIC_URI", publicHost)
	os.Setenv("MONGODB_REPLICA_SET", "tsuru")
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	defer func() {
		database := session().DB("myapp")
		database.RemoveUser("myapp")
		database.DropDatabase()
		collection().Remove(bson.M{"name": "myapp"})
	}()
	c.Assert(recorder.Code, gocheck.Equals, http.StatusCreated)
	var data map[string]string
	err = json.NewDecoder(recorder.Body).Decode(&data)
	c.Assert(err, gocheck.IsNil)
	c.Assert(data["MONGO_REPLICA_SET"], gocheck.Equals, "tsuru")
}

func (s *S) TestBindShouldReturnTheVariables(c *gocheck.C) {
	body := strings.NewReader("app-host=localhost&unit-host=127.0.0.1")
	request, err := http.NewRequest("POST", "/resources/myapp", body)
	c.Assert(err, gocheck.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	publicHost := "mongoapi.com:27017"
	os.Setenv("MONGODB_PUBLIC_URI", publicHost)
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	defer func() {
		database := session().DB("myapp")
		database.RemoveUser("myapp")
		database.DropDatabase()
		collection().Remove(bson.M{"name": "myapp"})
	}()
	c.Assert(recorder.Code, gocheck.Equals, http.StatusCreated)
	result, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, gocheck.IsNil)
	data := map[string]string{}
	json.Unmarshal(result, &data)
	c.Assert(data["MONGO_URI"], gocheck.Equals, publicHost)
	c.Assert(data["MONGO_USER"], gocheck.Equals, "myapp")
	c.Assert(data["MONGO_DATABASE_NAME"], gocheck.Equals, "myapp")
	c.Assert(data["MONGO_PASSWORD"], gocheck.Not(gocheck.HasLen), 0)
}

func (s *S) TestBindShouldCreateTheDatabase(c *gocheck.C) {
	body := strings.NewReader("app-host=localhost&unit-host=127.0.0.1")
	request, err := http.NewRequest("POST", "/resources/myapp", body)
	c.Assert(err, gocheck.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	defer func() {
		database := session().DB("myapp")
		database.RemoveUser("myapp")
		database.DropDatabase()
		collection().Remove(bson.M{"name": "myapp"})
	}()
	c.Assert(recorder.Code, gocheck.Equals, http.StatusCreated)
	databases, err := session().DatabaseNames()
	c.Assert("myapp", In, databases)
}

func (s *S) TestBindShouldAddUserInTheDatabase(c *gocheck.C) {
	body := strings.NewReader("app-host=localhost&unit-host=127.0.0.1")
	request, err := http.NewRequest("POST", "/resources/myapp", body)
	c.Assert(err, gocheck.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	defer func() {
		database := session().DB("myapp")
		database.RemoveUser("myapp")
		database.DropDatabase()
		collection().Remove(bson.M{"name": "myapp"})
	}()
	c.Assert(recorder.Code, gocheck.Equals, http.StatusCreated)
	collection := session().DB("myapp").C("system.users")
	lenght, err := collection.Find(bson.M{"user": "myapp"}).Count()
	c.Assert(lenght, gocheck.Equals, 1)
}

func (s *S) TestBindNoAppHost(c *gocheck.C) {
	body := strings.NewReader("unit-host=127.0.0.1")
	request, err := http.NewRequest("POST", "/resources/myapp", body)
	c.Assert(err, gocheck.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, gocheck.Equals, http.StatusBadRequest)
	c.Assert(recorder.Body.String(), gocheck.Equals, "Missing app-host")
}

func (s *S) TestBindNoUnitHost(c *gocheck.C) {
	body := strings.NewReader("app-host=localhost")
	request, err := http.NewRequest("POST", "/resources/myapp", body)
	c.Assert(err, gocheck.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, gocheck.Equals, http.StatusBadRequest)
	c.Assert(recorder.Body.String(), gocheck.Equals, "Missing unit-host")
}

func (s *S) TestUnbind(c *gocheck.C) {
	name := "myapp"
	_, err := bind(name, "localhost", "10.10.10.10")
	c.Assert(err, gocheck.IsNil)
	defer func() {
		database := session().DB(name)
		database.DropDatabase()
	}()
	request, err := http.NewRequest("DELETE", "/resources/myapp/hostname/10.10.10.10", nil)
	c.Assert(err, gocheck.IsNil)
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, gocheck.Equals, http.StatusOK)
	coll := session().DB(name).C("system.users")
	lenght, err := coll.Find(bson.M{"user": name}).Count()
	c.Assert(lenght, gocheck.Equals, 0)
	count, err := collection().Find(bson.M{"name": name}).Count()
	c.Assert(err, gocheck.IsNil)
	c.Assert(count, gocheck.Equals, 0)
}

func (s *S) TestUnbindWithoutRemovingTheUser(c *gocheck.C) {
	name := "myapp"
	_, err := bind(name, "localhost", "10.10.10.10")
	c.Assert(err, gocheck.IsNil)
	_, err = bind(name, "localhost", "10.10.10.11")
	c.Assert(err, gocheck.IsNil)
	request, err := http.NewRequest("DELETE", "/resources/myapp/hostname/10.10.10.10", nil)
	c.Assert(err, gocheck.IsNil)
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, gocheck.Equals, http.StatusOK)
	coll := session().DB(name).C("system.users")
	lenght, err := coll.Find(bson.M{"user": name}).Count()
	c.Assert(lenght, gocheck.Equals, 1)
	count, err := collection().Find(bson.M{"name": name}).Count()
	c.Assert(err, gocheck.IsNil)
	c.Assert(count, gocheck.Equals, 1)
}

func (s *S) TestRemoveShouldRemoveTheDatabase(c *gocheck.C) {
	name := "myapp"
	database := session().DB(name)
	database.AddUser(name, "", false)
	request, err := http.NewRequest("DELETE", "/resources/myapp", nil)
	c.Assert(err, gocheck.IsNil)
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, gocheck.Equals, http.StatusOK)
	databases, err := session().DatabaseNames()
	c.Assert(name, gocheck.Not(In), databases)
}

func (s *S) TestRemoveShouldRemoveBinds(c *gocheck.C) {
	name := "myapp"
	collection().Insert(dbBind{Name: name})
	database := session().DB(name)
	database.AddUser(name, "", false)
	request, err := http.NewRequest("DELETE", "/resources/myapp", nil)
	c.Assert(err, gocheck.IsNil)
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, gocheck.Equals, http.StatusOK)
	count, err := collection().Find(bson.M{"name": name}).Count()
	c.Assert(err, gocheck.IsNil)
	c.Assert(count, gocheck.Equals, 0)
}

func (s *S) TestStatus(c *gocheck.C) {
	name := "myapp"
	database := session().DB(name)
	database.AddUser(name, "", false)
	defer func() {
		database.RemoveUser("myapp")
		database.DropDatabase()
	}()
	request, err := http.NewRequest("GET", "/resources/myapp/status", nil)
	c.Assert(err, gocheck.IsNil)
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, gocheck.Equals, http.StatusNoContent)
}

func errorHandler(w http.ResponseWriter, r *http.Request) error {
	return stderrors.New("some error")
}

func httpErrorHandler(w http.ResponseWriter, r *http.Request) error {
	return &httpError{code: 400, body: "please provide a name"}
}

func simpleHandler(w http.ResponseWriter, r *http.Request) error {
	fmt.Fprint(w, "success")
	return nil
}

func (s *S) TestHandlerReturns500WhenInternalHandlerReturnsAnError(c *gocheck.C) {
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/apps", nil)
	c.Assert(err, gocheck.IsNil)
	Handler(errorHandler).ServeHTTP(recorder, request)
	c.Assert(recorder.Code, gocheck.Equals, 500)
	c.Assert(recorder.Body.String(), gocheck.Equals, "some error\n")
}

func (s *S) TestHandlerWithHTTPError(c *gocheck.C) {
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/apps", nil)
	c.Assert(err, gocheck.IsNil)
	Handler(httpErrorHandler).ServeHTTP(recorder, request)
	c.Assert(recorder.Code, gocheck.Equals, 400)
	c.Assert(recorder.Body.String(), gocheck.Equals, "please provide a name\n")
}

func (s *S) TestHandlerShouldPassAnHandlerWithoutError(c *gocheck.C) {
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/apps", nil)
	c.Assert(err, gocheck.IsNil)
	Handler(simpleHandler).ServeHTTP(recorder, request)
	c.Assert(recorder.Code, gocheck.Equals, 200)
	c.Assert(recorder.Body.String(), gocheck.Equals, "success")
}

func (s *S) TestHTTPError(c *gocheck.C) {
	var err error = &httpError{code: 404, body: "not found"}
	c.Assert(err.Error(), gocheck.Equals, "HTTP error (404): not found")
}
