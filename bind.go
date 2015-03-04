// Copyright 2015 mongoapi authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"

	"gopkg.in/mgo.v2"
)

// dbBind represents a bind stored in the database.
type dbBind struct {
	Name     string `bson:",omitempty"`
	AppHost  string `bson:",omitempty"`
	Password string `bson:",omitempty"`
}

type env map[string]string

var locker = multiLocker()

func bind(name, appHost string) (env, error) {
	data := map[string]string{
		"MONGO_URI":           coalesceEnv("MONGODB_PUBLIC_URI", "MONGODB_URI", "127.0.0.1:27017"),
		"MONGO_USER":          name,
		"MONGO_DATABASE_NAME": name,
	}
	locker.Lock(name)
	defer locker.Unlock(name)
	bind, err := newBind(name, appHost)
	if err != nil {
		return nil, err
	}
	data["MONGO_PASSWORD"] = bind.Password
	if rs := os.Getenv("MONGODB_REPLICA_SET"); rs != "" {
		data["MONGO_REPLICA_SET"] = rs
	}
	return env(data), nil
}

func newBind(name, appHost string) (dbBind, error) {
	password := newPassword()
	err := addUser(name, name, password)
	if err != nil {
		return dbBind{}, err
	}
	item := dbBind{AppHost: appHost, Name: name, Password: password}
	err = collection().Insert(item)
	if err != nil {
		return dbBind{}, err
	}
	return item, nil
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

func unbind(name, appHost string) error {
	locker.Lock(name)
	defer locker.Unlock(name)
	coll := collection()
	bind := dbBind{Name: name, AppHost: appHost}
	err := coll.Remove(bind)
	if err != nil {
		return err
	}
	return removeUser(name, name)
}

func removeUser(db, user string) error {
	database := session().DB(db)
	return database.RemoveUser(user)
}
