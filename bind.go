// Copyright 2013 mongoapi authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"labix.org/v2/mgo/bson"
	"os"
)

// dbBind represents a bind stored in the database.
type dbBind struct {
	AppHost  string
	Name     string
	Password string
	Units    []string
}

type env map[string]string

func bind(name, appHost, unitHost string) (env, error) {
	data := map[string]string{
		"MONGO_URI":           coalesceEnv("127.0.0.1:27017", "MONGODB_PUBLIC_URI", "MONGODB_URI"),
		"MONGO_USER":          name,
		"MONGO_DATABASE_NAME": name,
	}
	var bind dbBind
	q := bson.M{"name": name, "apphost": appHost}
	coll := collection()
	if err := coll.Find(q).One(&bind); err == nil {
		err = coll.Update(q, bson.M{"$addToSet": bson.M{"units": unitHost}})
		if err != nil {
			return nil, err
		}
		data["MONGO_PASSWORD"] = bind.Password
	} else if err.Error() != "not found" {
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
		Units:    []string{unitHost},
	})
	if err != nil {
		return "", err
	}
	return password, nil
}

func addUser(db, user, password string) error {
	database := session().DB(db)
	return database.AddUser(user, password, false)
}

func unbind(name, unitHost string) error {
	var bind dbBind
	coll := collection()
	q := bson.M{"name": name, "units": unitHost}
	err := coll.Find(q).One(&bind)
	if err != nil {
		return err
	}
	if len(bind.Units) == 1 {
		coll.Remove(q)
		return removeUser(name, name)
	}
	return coll.Update(q, bson.M{"$pull": bson.M{"units": unitHost}})
}

func removeUser(db, user string) error {
	database := session().DB(db)
	return database.RemoveUser(user)
}
