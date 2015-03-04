// Copyright 2015 mongoapi authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"

	"gopkg.in/check.v1"
)

func (s *S) TestSessionNoEnvVar(c *check.C) {
	os.Setenv("MONGODB_URI", "")
	session := session()
	servers := session.LiveServers()
	c.Assert(servers, check.DeepEquals, []string{"127.0.0.1:27017"})
}

func (s *S) TestSessionDontConnectTwice(c *check.C) {
	os.Setenv("MONGODB_URI", "")
	session1 := session()
	session2 := session()
	c.Assert(session1, check.Equals, session2)
}

func (s *S) TestSessionReconnects(c *check.C) {
	session1 := session()
	session1.Close()
	session2 := session()
	err := session2.Ping()
	c.Assert(err, check.IsNil)
}

func (s *S) TestSessionUsesEnvironmentVariable(c *check.C) {
	os.Setenv("MONGODB_URI", "localhost:27017")
	sess.Close()
	session := session()
	servers := session.LiveServers()
	c.Assert(servers, check.DeepEquals, []string{"localhost:27017"})
}

func (s *S) TestCoalesceEnv(c *check.C) {
	var tests = []struct {
		envs []string
		want string
	}{
		{
			envs: []string{"HOSAOJSWJMDKDKD", "88isosuukkd", "default"},
			want: "default",
		},
		{
			envs: []string{"HOME", "PATH", "default"},
			want: os.Getenv("HOME"),
		},
	}
	for _, t := range tests {
		c.Check(coalesceEnv(t.envs...), check.Equals, t.want)
	}
}

func (s *S) TestDBNameDefaultValue(c *check.C) {
	c.Assert(dbName(), check.Equals, "mongoapi")
}

func (s *S) TestDBNameEnvVar(c *check.C) {
	os.Setenv("MONGOAPI_DBNAME", "mongo_api")
	defer os.Setenv("MONGOAPI_DBNAME", "")
	c.Assert(dbName(), check.Equals, "mongo_api")
}

func (s *S) TestCollection(c *check.C) {
	coll := collection()
	c.Assert(coll.Database.Name, check.Equals, "mongoapi")
	err := coll.Database.Session.Ping()
	c.Assert(err, check.IsNil)
}
