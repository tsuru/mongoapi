// Copyright 2013 mongoapi authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"labix.org/v2/mgo"
	"os"
)

var sess *mgo.Session

func session() (s *mgo.Session) {
	var connect = func() *mgo.Session {
		var err error
		uri := coalesceEnv("127.0.0.1:27017", "MONGODB_URI")
		session, err := mgo.Dial(uri)
		if err != nil {
			panic(err)
		}
		return session
	}
	if sess == nil {
		sess = connect()
	}
	defer func() {
		if r := recover(); r != nil {
			sess = connect()
			s = sess
		}
	}()
	err := sess.Ping()
	if err != nil {
		sess = connect()
	}
	return sess
}

// coalesceEnv returns the value of the first environment variable in the list
// that is not empty, or the default value.
func coalesceEnv(defaultValue string, envs ...string) string {
	for _, e := range envs {
		if value := os.Getenv(e); value != "" {
			return value
		}
	}
	return defaultValue
}

func dbName() string {
	return coalesceEnv("mongoapi", "MONGOAPI_DBNAME")
}
