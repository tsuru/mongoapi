package main

import "net/http"

func AddInstance(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}
