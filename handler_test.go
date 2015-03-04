// Copyright 2015 mongoapi authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"gopkg.in/check.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var _ = check.Suite(&S{})

type S struct {
	muxer http.Handler
}

func Test(t *testing.T) { check.TestingT(t) }

func (s *S) SetUpSuite(c *check.C) {
	s.muxer = buildMux()
}

func (s *S) TearDownSuite(c *check.C) {
	session().DB(dbName()).DropDatabase()
}

func (s *S) SetUpTest(c *check.C) {
	collection().RemoveAll(nil)
}

type InChecker struct{}

func (c *InChecker) Info() *check.CheckerInfo {
	return &check.CheckerInfo{Name: "In", Params: []string{"value", "list"}}
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

var In check.Checker = &InChecker{}

func (s *S) TestAdd(c *check.C) {
	body := strings.NewReader("name=something")
	request, err := http.NewRequest("POST", "/resources", body)
	c.Assert(err, check.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusCreated)
}

func (s *S) TestAddReservedName(c *check.C) {
	name := dbName()
	body := strings.NewReader("name=" + name)
	request, err := http.NewRequest("POST", "/resources", body)
	c.Assert(err, check.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusForbidden)
	c.Assert(recorder.Body.String(), check.Equals, "Reserved name")
}

func (s *S) TestBindShouldReturnLocalhostWhenThePublicHostEnvIsNil(c *check.C) {
	body := strings.NewReader("app-host=localhost&unit-host=127.0.0.1")
	request, err := http.NewRequest("POST", "/resources/myapp", body)
	c.Assert(err, check.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	os.Setenv("MONGODB_PUBLIC_URI", "")
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	defer func() {
		database := session().DB("myapp")
		database.RemoveUser("myapp")
		database.DropDatabase()
	}()
	c.Assert(recorder.Code, check.Equals, http.StatusCreated)
	result, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, check.IsNil)
	data := map[string]string{}
	json.Unmarshal(result, &data)
	c.Assert(data["MONGO_URI"], check.Equals, "127.0.0.1:27017")
	c.Assert(data["MONGO_USER"], check.Equals, "myapp")
	c.Assert(data["MONGO_DATABASE_NAME"], check.Equals, "myapp")
	c.Assert(data["MONGO_PASSWORD"], check.Not(check.HasLen), 0)
	coll := collection()
	expected := dbBind{
		AppHost:  "localhost",
		Name:     "myapp",
		Password: data["MONGO_PASSWORD"],
		Calls:    1,
	}
	var bind dbBind
	q := bson.M{"name": "myapp"}
	defer coll.Remove(q)
	err = coll.Find(q).One(&bind)
	c.Assert(err, check.IsNil)
	c.Assert(bind, check.DeepEquals, expected)
}

func (s *S) TestBindTwoInstances(c *check.C) {
	body := strings.NewReader("app-host=localhost&unit-host=127.0.0.1")
	request, err := http.NewRequest("POST", "/resources/myapp?:name=myapp", body)
	c.Assert(err, check.IsNil)
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
	c.Assert(err, check.IsNil)
	body = strings.NewReader("app-host=localhost&unit-host=127.0.0.2")
	request, err = http.NewRequest("POST", "/resources/myapp", body)
	c.Assert(err, check.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder = httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusCreated)
	err = json.NewDecoder(recorder.Body).Decode(&second)
	c.Assert(err, check.IsNil)
	c.Assert(second, check.DeepEquals, first)
	expected := dbBind{
		AppHost:  "localhost",
		Name:     "myapp",
		Password: first["MONGO_PASSWORD"],
		Calls:    2,
	}
	var bind dbBind
	err = collection().Find(bson.M{"name": "myapp"}).One(&bind)
	c.Assert(err, check.IsNil)
	c.Assert(bind, check.DeepEquals, expected)
}

func (s *S) TestBindWithReplicaSet(c *check.C) {
	body := strings.NewReader("app-host=localhost&unit-host=127.0.0.1")
	request, err := http.NewRequest("POST", "/resources/myapp", body)
	c.Assert(err, check.IsNil)
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
	c.Assert(recorder.Code, check.Equals, http.StatusCreated)
	var data map[string]string
	err = json.NewDecoder(recorder.Body).Decode(&data)
	c.Assert(err, check.IsNil)
	c.Assert(data["MONGO_REPLICA_SET"], check.Equals, "tsuru")
}

func (s *S) TestBind(c *check.C) {
	body := strings.NewReader("app-host=localhost&unit-host=127.0.0.1")
	request, err := http.NewRequest("POST", "/resources/myapp", body)
	c.Assert(err, check.IsNil)
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
	c.Assert(recorder.Code, check.Equals, http.StatusCreated)
	result, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, check.IsNil)
	data := map[string]string{}
	json.Unmarshal(result, &data)
	c.Assert(data["MONGO_URI"], check.Equals, publicHost)
	c.Assert(data["MONGO_USER"], check.Equals, "myapp")
	c.Assert(data["MONGO_DATABASE_NAME"], check.Equals, "myapp")
	c.Assert(data["MONGO_PASSWORD"], check.Not(check.HasLen), 0)
	info := mgo.DialInfo{
		Addrs:    []string{"localhost:27017"},
		Database: data["MONGO_DATABASE_NAME"],
		Username: data["MONGO_USER"],
		Password: data["MONGO_PASSWORD"],
	}
	session, err := mgo.DialWithInfo(&info)
	c.Assert(err, check.IsNil)
	defer session.Close()
	err = session.DB(info.Database).C("mycollection").Insert(bson.M{"some": "stuff"})
	c.Assert(err, check.IsNil)
	err = session.DB(info.Database).C("mycollection").Remove(bson.M{"some": "stuff"})
}

func (s *S) TestBindNoAppHost(c *check.C) {
	body := strings.NewReader("unit-host=127.0.0.1")
	request, err := http.NewRequest("POST", "/resources/myapp", body)
	c.Assert(err, check.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
	c.Assert(recorder.Body.String(), check.Equals, "Missing app-host")
}

func (s *S) TestBindNoUnitHost(c *check.C) {
	body := strings.NewReader("app-host=localhost")
	request, err := http.NewRequest("POST", "/resources/myapp", body)
	c.Assert(err, check.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
	c.Assert(recorder.Body.String(), check.Equals, "Missing unit-host")
}

func (s *S) TestUnbind(c *check.C) {
	name := "myapp"
	env, err := bind(name, "localhost", "10.10.10.10")
	c.Assert(err, check.IsNil)
	defer func() {
		database := session().DB(name)
		database.DropDatabase()
	}()
	request, err := http.NewRequest("DELETE", "/resources/myapp/hostname/10.10.10.10", nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	info := mgo.DialInfo{
		Addrs:    []string{"localhost:27017"},
		Database: env["MONGO_DATABASE"],
		Username: env["MONGO_USER"],
		Password: env["MONGO_PASSWORD"],
		Timeout:  1e9,
		FailFast: true,
	}
	_, err = mgo.DialWithInfo(&info)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "auth failed")
}

func (s *S) TestUnbindWithoutRemovingTheUser(c *check.C) {
	name := "myapp"
	env, err := bind(name, "localhost", "10.10.10.10")
	c.Assert(err, check.IsNil)
	_, err = bind(name, "localhost", "10.10.10.11")
	c.Assert(err, check.IsNil)
	request, err := http.NewRequest("DELETE", "/resources/myapp/hostname/10.10.10.10", nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	info := mgo.DialInfo{
		Addrs:    []string{"localhost:27017"},
		Database: env["MONGO_DATABASE_NAME"],
		Username: env["MONGO_USER"],
		Password: env["MONGO_PASSWORD"],
		Timeout:  1e9,
		FailFast: true,
	}
	session, err := mgo.DialWithInfo(&info)
	c.Assert(err, check.IsNil)
	session.Close()
	count, err := collection().Find(bson.M{"name": name}).Count()
	c.Assert(err, check.IsNil)
	c.Assert(count, check.Equals, 1)
}

func (s *S) TestRemoveShouldRemoveTheDatabase(c *check.C) {
	name := "myapp"
	database := session().DB(name)
	database.AddUser(name, "", false)
	request, err := http.NewRequest("DELETE", "/resources/myapp", nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	databases, err := session().DatabaseNames()
	c.Assert(name, check.Not(In), databases)
}

func (s *S) TestRemoveShouldRemoveBinds(c *check.C) {
	name := "myapp"
	collection().Insert(dbBind{Name: name})
	database := session().DB(name)
	database.AddUser(name, "", false)
	request, err := http.NewRequest("DELETE", "/resources/myapp", nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	count, err := collection().Find(bson.M{"name": name}).Count()
	c.Assert(err, check.IsNil)
	c.Assert(count, check.Equals, 0)
}

func (s *S) TestStatus(c *check.C) {
	name := "myapp"
	database := session().DB(name)
	database.AddUser(name, "", false)
	defer func() {
		database.RemoveUser("myapp")
		database.DropDatabase()
	}()
	request, err := http.NewRequest("GET", "/resources/myapp/status", nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNoContent)
}

func errorHandler(w http.ResponseWriter, r *http.Request) error {
	return errors.New("some error")
}

func httpErrorHandler(w http.ResponseWriter, r *http.Request) error {
	return &httpError{code: 400, body: "please provide a name"}
}

func simpleHandler(w http.ResponseWriter, r *http.Request) error {
	fmt.Fprint(w, "success")
	return nil
}

func (s *S) TestHandlerReturns500WhenInternalHandlerReturnsAnError(c *check.C) {
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/apps", nil)
	c.Assert(err, check.IsNil)
	Handler(errorHandler).ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, 500)
	c.Assert(recorder.Body.String(), check.Equals, "some error\n")
}

func (s *S) TestHandlerWithHTTPError(c *check.C) {
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/apps", nil)
	c.Assert(err, check.IsNil)
	Handler(httpErrorHandler).ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, 400)
	c.Assert(recorder.Body.String(), check.Equals, "please provide a name\n")
}

func (s *S) TestHandlerShouldPassAnHandlerWithoutError(c *check.C) {
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/apps", nil)
	c.Assert(err, check.IsNil)
	Handler(simpleHandler).ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, 200)
	c.Assert(recorder.Body.String(), check.Equals, "success")
}

func (s *S) TestHTTPError(c *check.C) {
	var err error = &httpError{code: 404, body: "not found"}
	c.Assert(err.Error(), check.Equals, "HTTP error (404): not found")
}
