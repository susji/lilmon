package main

import (
	"database/sql"
	"flag"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func db_init(db_path string) *sql.DB {
	db, err := sql.Open("sqlite", db_path)
	if err != nil {
		log.Fatal("cannot open database: %v", err)
	}

	var db_version string
	if err := db.QueryRow("select sqlite_version()").Scan(&db_version); err != nil {
		log.Println("warning: unable to get sqlite version: ", err)
	} else {
		log.Println("database version: ", db_version)
	}
	return db
}

func main() {
	var db_path string

	flag.StringVar(&db_path, "db-path", "/tmp/tinmon.sqlite", "Path where to store tinmon database file")
	flag.Parse()

	db := db_init(db_path)
	defer func() {
		if err := db.Close(); err != nil {
			log.Println("warning: error when closing database: ", err)
		}
	}()

}
