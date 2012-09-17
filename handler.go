package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Add(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func Bind(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"MONGO_URI":           "127.0.0.1:27017",
		"MONGO_USER":          "myapp",
		"MONGO_PASSWORD":      "",
		"MONGO_DATABASE_NAME": r.URL.Query().Get(":name"),
	}
	b, _ := json.Marshal(&data)
	fmt.Fprint(w, string(b))
	w.WriteHeader(http.StatusCreated)
}

func Unbind(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func Remove(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
