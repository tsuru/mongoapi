// Copyright 2013 mongoapi authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"launchpad.net/gocheck"
	"os"
)

func (s *S) TestSessionNoEnvVar(c *gocheck.C) {
	os.Setenv("MONGODB_URI", "")
	session := session()
	servers := session.LiveServers()
	c.Assert(servers, gocheck.DeepEquals, []string{"127.0.0.1:27017"})
}

func (s *S) TestSessionDontConnectTwice(c *gocheck.C) {
	os.Setenv("MONGODB_URI", "")
	session1 := session()
	session2 := session()
	c.Assert(session1, gocheck.Equals, session2)
}

func (s *S) TestSessionReconnects(c *gocheck.C) {
	session1 := session()
	session1.Close()
	session2 := session()
	err := session2.Ping()
	c.Assert(err, gocheck.IsNil)
}

func (s *S) TestSessionUsesEnvironmentVariable(c *gocheck.C) {
	os.Setenv("MONGODB_URI", "localhost:27017")
	sess.Close()
	session := session()
	servers := session.LiveServers()
	c.Assert(servers, gocheck.DeepEquals, []string{"localhost:27017"})
}

func (s *S) TestCoalesceEnv(c *gocheck.C) {
	var tests = []struct {
		dvalue string
		envs   []string
		want   string
	}{
		{
			dvalue: "default",
			envs:   []string{"HOSAOJSWJMDKDKD", "88isosuukkd"},
			want:   "default",
		},
		{
			dvalue: "",
			envs:   []string{"HOME", "PATH"},
			want:   os.Getenv("HOME"),
		},
	}
	for _, t := range tests {
		c.Check(coalesceEnv(t.dvalue, t.envs...), gocheck.Equals, t.want)
	}
}

func (s *S) TestDBNameDefaultValue(c *gocheck.C) {
	c.Assert(dbName(), gocheck.Equals, "mongoapi")
}

func (s *S) TestDBNameEnvVar(c *gocheck.C) {
	os.Setenv("MONGOAPI_DBNAME", "mongo_api")
	defer os.Setenv("MONGOAPI_DBNAME", "")
	c.Assert(dbName(), gocheck.Equals, "mongo_api")
}
