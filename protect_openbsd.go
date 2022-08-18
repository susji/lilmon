// +build openbsd

package main

import (
	"log"
	"os"

	"golang.org/x/sys/unix"
)

const (
	promises_serve     = "inet stdio rpath wpath cpath tmppath flock dns"
	execpromises_serve = ""

	promises_measure = "stdio proc exec flock tmppath"

	// `serve` may not need `c` for the database directory with SQLite, but
	// we'll give it just in case. It may be that WAL maintenance requires
	// this even though `serve` is using read-only access.
	unveilflags_db  = "rwc"
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
	log.Printf("pledge: promises=%q, execpromises=%q\n", promises_serve, execpromises_serve)
	if err := unix.Pledge(promises_serve, execpromises_serve); err != nil {
		return err
	}
	return nil
}

func protect_measure() error {
	// measure is hard to unveil, because it has to be compatible with a
	// very rich selection of shell commands.
	log.Printf("pledge: promises=%q\n", promises_measure)
	if err := unix.PledgePromises(promises_measure); err != nil {
		return err
	}
	return nil
}
