// Copyright 2015 mongoapi authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// dbBind represents a bind stored in the database.
type dbBind struct {
	Name     string
	AppHost  string
	Password string
	Calls    int
}

type env map[string]string

var locker = multiLocker()

func bind(name, appHost, unitHost string) (env, error) {
	data := map[string]string{
		"MONGO_URI":           coalesceEnv("127.0.0.1:27017", "MONGODB_PUBLIC_URI", "MONGODB_URI"),
		"MONGO_USER":          name,
		"MONGO_DATABASE_NAME": name,
	}
	locker.Lock(name)
	defer locker.Unlock(name)
	var bind dbBind
	q := bson.M{"name": name, "apphost": appHost}
	coll := collection()
	if err := coll.Find(q).One(&bind); err == nil {
		err = coll.Update(q, bson.M{"$inc": bson.M{"calls": 1}})
		if err != nil {
			return nil, err
		}
		data["MONGO_PASSWORD"] = bind.Password
	} else if err != mgo.ErrNotFound {
		return nil, err
	} else {
		password, err := newBind(name, appHost, unitHost)
		if err != nil {
			return nil, err
		}
		data["MONGO_PASSWORD"] = password
	}
	if rs := os.Getenv("MONGODB_REPLICA_SET"); rs != "" {
		data["MONGO_REPLICA_SET"] = rs
	}
	return env(data), nil
}

func newBind(name, appHost, unitHost string) (string, error) {
	password := newPassword()
	err := addUser(name, name, password)
	if err != nil {
		return "", err
	}
	err = collection().Insert(dbBind{
		AppHost:  appHost,
		Name:     name,
		Password: password,
		Calls:    1,
	})
	if err != nil {
		return "", err
	}
	return password, nil
}

func addUser(db, username, password string) error {
	database := session().DB(db)
	user := mgo.User{
		Username: username,
		Password: password,
		Roles:    []mgo.Role{mgo.RoleReadWrite},
	}
	return database.UpsertUser(&user)
}

func unbind(name string) error {
	locker.Lock(name)
	defer locker.Unlock(name)
	var bind dbBind
	coll := collection()
	q := bson.M{"name": name}
	err := coll.Find(q).One(&bind)
	if err != nil {
		return err
	}
	if bind.Calls == 1 {
		coll.Remove(q)
		return removeUser(name, name)
	}
	return coll.Update(q, bson.M{"$inc": bson.M{"calls": -1}})
}

func removeUser(db, user string) error {
	database := session().DB(db)
	return database.RemoveUser(user)
}
