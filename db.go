// Copyright 2015 mongoapi authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"

	"gopkg.in/mgo.v2"
)

var sess *mgo.Session

func session() (s *mgo.Session) {
	var connect = func() *mgo.Session {
		var err error
		uri := coalesceEnv("MONGODB_URI", "127.0.0.1:27017")
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
func coalesceEnv(envs ...string) string {
	if len(envs) == 0 {
		return ""
	}
	length := len(envs)
	defaultValue := envs[length-1]
	for _, e := range envs[:length-1] {
		if value := os.Getenv(e); value != "" {
			return value
		}
	}
	return defaultValue
}

func dbName() string {
	return coalesceEnv("MONGOAPI_DBNAME", "mongoapi")
}

func collection() *mgo.Collection {
	return session().DB(dbName()).C("bind")
}
