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

func (s *S) TestAdd(c *C) {
	request, err := http.NewRequest("POST", "/resources/", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	Add(recorder, request)
	c.Assert(recorder.Code, Equals, http.StatusCreated)
}

func (s *S) TestBind(c *C) {
	request, err := http.NewRequest("POST", "/resources/myapp?:name=myapp", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	Bind(recorder, request)
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

func (s *S) TestUnbind(c *C) {
	request, err := http.NewRequest("DELETE", "/resources/myapp/hostname/10.10.10.10?:name=myapp&hostname=10.10.10.10", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	Unbind(recorder, request)
	c.Assert(recorder.Code, Equals, http.StatusOK)
}

func (s *S) TestRemove(c *C) {
	request, err := http.NewRequest("DELETE", "/resources/name?:name=name", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	Remove(recorder, request)
	c.Assert(recorder.Code, Equals, http.StatusOK)
}

func (s *S) TestStatus(c *C) {
	request, err := http.NewRequest("GET", "/resources/myapp/status?:name=myapp", nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	Status(recorder, request)
	c.Assert(recorder.Code, Equals, http.StatusNoContent)
}
