package main

import (
	"github.com/bmizerany/pat"
	"log"
	"net/http"
)

func main() {
	m := pat.New()
	m.Post("/resources", http.HandlerFunc(Add))
	m.Post("/resources/:name", Handler(Bind))
	m.Del("/resources/:name/hostname/:hostname", Handler(Unbind))
	m.Del("/resources/:name", Handler(Remove))
	m.Get("/resources/:name/status", Handler(Status))
	log.Fatal(http.ListenAndServe(":3333", m))
}
