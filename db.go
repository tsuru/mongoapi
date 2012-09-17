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
