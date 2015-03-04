// Copyright 2015 mongoapi authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"net/http"

	"gopkg.in/mgo.v2/bson"
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
	err := unbind(name)
	if err == nil {
		w.WriteHeader(http.StatusOK)
	}
	return err
}

func Remove(w http.ResponseWriter, r *http.Request) error {
	name := r.URL.Query().Get(":name")
	collection().RemoveAll(bson.M{"name": name})
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
		if e, ok := err.(*httpError); ok {
			http.Error(w, e.body, e.code)
		} else {
			http.Error(w, err.Error(), 500)
		}
	}
}

type httpError struct {
	code int
	body string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("HTTP error (%d): %s", e.code, e.body)
}
