// +build openbsd

package main

import (
	"log"
	"os"

	"golang.org/x/sys/unix"
)

const (
	promises        = "inet stdio rpath wpath cpath tmppath flock dns"
	execpromises    = ""
	unveilflags_db  = "rw"
	unveilflags_tmp = "rwc"
)

func protect_serve(path_db string) error {
	log.Printf("unveil database directory: path=%q, flags=%q\n", path_db, unveilflags_db)
	if err := unix.Unveil(path_db, unveilflags_db); err != nil {
		return err
	}
	log.Printf("unveil temp directory: path=%q, flags=%q\n", os.TempDir(), unveilflags_tmp)
	if err := unix.Unveil(os.TempDir(), unveilflags_tmp); err != nil {
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
