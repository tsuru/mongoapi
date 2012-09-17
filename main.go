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
	m.Del("/resources/:name", http.HandlerFunc(Remove))
	log.Fatal(http.ListenAndServe(":3333", m))
}
