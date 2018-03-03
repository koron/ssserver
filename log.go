package main

import (
	"log"
	"os"
)

var (
	infoL *log.Logger
	warnL = log.New(os.Stderr, "WARN ", log.LstdFlags)
)

func infof(s string, args ...interface{}) {
	if infoL == nil {
		return
	}
	infoL.Printf(s, args...)
}

func warnf(s string, args ...interface{}) {
	if warnL == nil {
		return
	}
	warnL.Printf(s, args...)
}
