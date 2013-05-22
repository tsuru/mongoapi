// Copyright 2013 mongoapi authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"labix.org/v2/mgo/bson"
	"net/http"
)

func newPassword() string {
	var random [32]byte
	rand.Read(random[:])
	h := sha512.New()
	h.Sum([]byte(bson.NewObjectId().Hex()))
	h.Sum(random[:])
	return fmt.Sprintf("%x", h.Sum(nil))
}

func Add(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func Bind(w http.ResponseWriter, r *http.Request) error {
	name := r.URL.Query().Get(":name")
	database := session().DB(name)
	err := database.AddUser(name, "", false)
	if err != nil {
		return err
	}
	data := map[string]string{
		"MONGO_URI":           coalesceEnv("127.0.0.1:27017", "MONGODB_PUBLIC_URI", "MONGODB_URI"),
		"MONGO_USER":          name,
		"MONGO_PASSWORD":      newPassword(),
		"MONGO_DATABASE_NAME": name,
	}
	b, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(b)
	return err
}

func Unbind(w http.ResponseWriter, r *http.Request) error {
	name := r.URL.Query().Get(":name")
	database := session().DB(name)
	err := database.RemoveUser(name)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	return nil
}

func Remove(w http.ResponseWriter, r *http.Request) error {
	name := r.URL.Query().Get(":name")
	err := session().DB(name).DropDatabase()
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	return nil
}

func Status(w http.ResponseWriter, r *http.Request) error {
	if err := session().Ping(); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

type Handler func(http.ResponseWriter, *http.Request) error

func (fn Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
