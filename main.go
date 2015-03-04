// Copyright 2015 mongoapi authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/bmizerany/pat"
)

const version = "0.2.0"

var printVersion bool
var listen string

func init() {
	flag.BoolVar(&printVersion, "v", false, "Print version and exit")
	flag.StringVar(&listen, "bind", "0.0.0.0:3030", "Bind the service on this port")
	flag.Parse()
}

func buildMux() http.Handler {
	m := pat.New()
	m.Post("/resources", http.HandlerFunc(Add))
	m.Post("/resources/:name/bind-app", Handler(BindApp))
	m.Del("/resources/:name/bind-app", Handler(UnbindApp))
	m.Post("/resources/:name/bind", Handler(BindUnit))
	m.Del("/resources/:name/bind", Handler(UnbindUnit))
	m.Del("/resources/:name", Handler(Remove))
	m.Get("/resources/:name/status", Handler(Status))
	return m
}

func main() {
	if printVersion {
		fmt.Printf("mongoapi version %s", version)
		return
	}
	log.Fatal(http.ListenAndServe(listen, buildMux()))
}
