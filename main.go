// Copyright 2013 mongoapi authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/bmizerany/pat"
	"log"
	"net/http"
)

const version = "0.1"

var printVersion bool

func init() {
	flag.BoolVar(&printVersion, "v", false, "Print version and exit")
	flag.Parse()
}

func buildMux() *pat.PatternServeMux {
	m := pat.New()
	m.Post("/resources", http.HandlerFunc(Add))
	m.Post("/resources/:name", Handler(Bind))
	m.Del("/resources/:name/hostname/:hostname", Handler(Unbind))
	m.Del("/resources/:name", Handler(Remove))
	m.Get("/resources/:name/status", Handler(Status))
	return m
}

func main() {
	if printVersion {
		fmt.Printf("mongoapi version %s", version)
		return
	}
	listen := coalesceEnv("0.0.0.0:3333", "MONGODB_API_LISTEN")
	log.Fatal(http.ListenAndServe(listen, buildMux()))
}
