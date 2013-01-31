// Copyright 2013 mongoapi authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "labix.org/v2/mgo"

var Session *mgo.Session

func init() {
	var err error
	Session, err = mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
}
