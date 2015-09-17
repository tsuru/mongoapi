// Copyright 2015 mongoapi authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"os"

	"gopkg.in/mgo.v2"
)

// dbBind represents a bind stored in the database.
type dbBind struct {
	Name     string `bson:",omitempty"`
	User     string `bson:",omitempty"`
	AppHost  string `bson:",omitempty"`
	Password string `bson:",omitempty"`
}

type env map[string]string

var locker = multiLocker()

func bind(name, appHost string) (env, error) {
	locker.Lock(name)
	defer locker.Unlock(name)
	bind, err := newBind(name, appHost)
	if err != nil {
		return nil, err
	}
	hosts := coalesceEnv("MONGODB_PUBLIC_URI", "MONGODB_URI", "127.0.0.1:27017")
	data := map[string]string{
		"MONGODB_HOSTS":         hosts,
		"MONGODB_USER":          bind.User,
		"MONGODB_PASSWORD":      bind.Password,
		"MONGODB_DATABASE_NAME": name,
	}
	var connStringSuffix string
	if rs := os.Getenv("MONGODB_REPLICA_SET"); rs != "" {
		data["MONGODB_REPLICA_SET"] = rs
		connStringSuffix = "?replicaSet=" + rs
	}
	data["MONGODB_CONNECTION_STRING"] = fmt.Sprintf(
		"mongodb://%s:%s@%s/%s%s",
		bind.User, bind.Password, hosts, name, connStringSuffix,
	)
	return env(data), nil
}

func newBind(name, appHost string) (dbBind, error) {
	password := newPassword()
	username := name + newPassword()[:8]
	err := addUser(name, username, password)
	if err != nil {
		return dbBind{}, err
	}
	item := dbBind{AppHost: appHost, User: username, Name: name, Password: password}
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
	err := coll.Find(bind).One(&bind)
	if err != nil {
		return err
	}
	err = coll.Remove(bind)
	if err != nil {
		return err
	}
	return removeUser(name, bind.User)
}

func removeUser(db, user string) error {
	database := session().DB(db)
	return database.RemoveUser(user)
}

func newPassword() string {
	var random [32]byte
	rand.Read(random[:])
	h := sha512.New()
	return fmt.Sprintf("%x", h.Sum(random[:]))
}
