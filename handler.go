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
	name := r.FormValue("name")
	if name == dbName() {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Reserved name")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func Bind(w http.ResponseWriter, r *http.Request) error {
	name := r.URL.Query().Get(":name")
	appHost := r.FormValue("app-host")
	if appHost == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Missing app-host")
		return nil
	}
	unitHost := r.FormValue("unit-host")
	if unitHost == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Missing unit-host")
		return nil
	}
	env, err := bind(name, appHost, unitHost)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(env)
}

func Unbind(w http.ResponseWriter, r *http.Request) error {
	name := r.URL.Query().Get(":name")
	host := r.URL.Query().Get(":hostname")
	err := unbind(name, host)
	if err == nil {
		w.WriteHeader(http.StatusOK)
	}
	return err
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
