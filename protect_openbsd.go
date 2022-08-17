// +build openbsd

package main

import (
	"log"

	"golang.org/x/sys/unix"
)

const (
	promises     = "inet stdio rpath wpath cpath flock dns"
	execpromises = ""
	unveilflags  = "rw"
)

func protect_serve(path_db string) error {
	log.Printf("unveil: path=%q, flags=%q\n", path_db, unveilflags)
	if err := unix.Unveil(path_db, unveilflags); err != nil {
		return err
	}
	if err := unix.UnveilBlock(); err != nil {
		return err
	}
	log.Printf("pledge: promises=%q, execpromises=%q\n", promises, execpromises)
	if err := unix.Pledge(promises, execpromises); err != nil {
		return err
	}
	return nil
}

func protect_measure() error {
	return nil
}
