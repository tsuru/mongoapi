package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func AddInstance(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func BindInstance(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"MONGO_URI":           "127.0.0.1:27017",
		"MONGO_USER":          "",
		"MONGO_PASSWORD":      "",
		"MONGO_DATABASE_NAME": r.URL.Query().Get(":name"),
	}
	b, _ := json.Marshal(&data)
	fmt.Fprint(w, string(b))
	w.WriteHeader(http.StatusCreated)
}

func RemoveInstance(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
