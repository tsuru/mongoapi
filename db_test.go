package main

import (
	"labix.org/v2/mgo"
	. "launchpad.net/gocheck"
)

func (s *S) TestSessionShouldReturnAMongoSession(c *C) {
	var session interface{}
	session = Session
	_, ok := session.(*mgo.Session)
	c.Assert(ok, Equals, true)
}
