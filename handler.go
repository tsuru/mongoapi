package main

import (
	"encoding/json"
	"fmt"
	"labix.org/v2/mgo"
	"net/http"
)

func Add(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func Bind(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(":name")
	database := Session.DB(name)
	database.AddUser(name, "", false)
	data := map[string]string{
		"MONGO_URI":           "127.0.0.1:27017",
		"MONGO_USER":          name,
		"MONGO_PASSWORD":      "",
		"MONGO_DATABASE_NAME": name,
	}
	b, _ := json.Marshal(&data)
	fmt.Fprint(w, string(b))
	w.WriteHeader(http.StatusCreated)
}

func Unbind(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(":name")
	database := Session.DB(name)
	database.RemoveUser(name)
	w.WriteHeader(http.StatusOK)
}

func Remove(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(":name")
	Session.DB(name).DropDatabase()
	w.WriteHeader(http.StatusOK)
}

func Status(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(":name")
	uri := fmt.Sprintf("%s:@127.0.0.1:27017/%s", name, name)
	_, err := mgo.Dial(uri)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type Handler func(http.ResponseWriter, *http.Request) error

func (fn Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
