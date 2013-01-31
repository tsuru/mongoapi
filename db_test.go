// Copyright 2013 mongoapi authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
