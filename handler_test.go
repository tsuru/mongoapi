package main

import (
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
