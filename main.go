package main

import (
	"github.com/bmizerany/pat"
	"log"
	"net/http"
)

func main() {
	m := pat.New()
	m.Post("/resources", http.HandlerFunc(Add))
	m.Post("/resources/:name", http.HandlerFunc(Bind))
	m.Del("/resources/:name/hostname/:hostname", http.HandlerFunc(Unbind))
	m.Del("/resources/:name", http.HandlerFunc(Remove))
	m.Get("/resources/:name/status", http.HandlerFunc(Status))
	log.Fatal(http.ListenAndServe(":3333", m))
}
