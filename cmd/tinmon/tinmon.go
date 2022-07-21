package main

import (
	"database/sql"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	db_filename := "/tmp/tinmon.sqlite"
	db, err := sql.Open("sqlite", db_filename)
	if err != nil {
		log.Fatal("cannot open database: %v", err)
	}

	var db_version string
	if err := db.QueryRow("select sqlite_version()").Scan(&db_version); err != nil {
		log.Println("warning: unable to get sqlite version: ", err)
	} else {
		log.Println("database version: ", db_version)
	}
}
