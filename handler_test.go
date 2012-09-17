package main

import (
	"encoding/json"
	"io/ioutil"
	. "launchpad.net/gocheck"
	"net/http"
	"net/http/httptest"
	"testing"
)

var _ = Suite(&S{})

type S struct{}

func Test(t *testing.T) { TestingT(t) }

func (s *S) TestAddInstance(c *C) {
	request, err := http.NewRequest("POST", "/resources/", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	AddInstance(recorder, request)
	c.Assert(recorder.Code, Equals, http.StatusCreated)
}

func (s *S) TestBindInstance(c *C) {
	request, err := http.NewRequest("POST", "/resources/myapp?:name=myapp", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	BindInstance(recorder, request)
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

func (s *S) TestRemoveInstance(c *C) {
	request, err := http.NewRequest("DELETE", "/resources/name?:name=name", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	RemoveInstance(recorder, request)
	c.Assert(recorder.Code, Equals, http.StatusOK)
}
