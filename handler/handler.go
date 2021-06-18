package handler

import "gopkg.in/mgo.v2"

type Handler struct {
	DB *mgo.Session
}

const (
	// Key (should come from somewhere else).
	Key = "secret"
)
