package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
)

func measure(path_config string) {
	config, err := config_load_file(path_config)
	if err != nil {
		log.Fatal(err)
	}

	mconfig, err := config.parse_measure()
	if err != nil {
		log.Fatal(err)
	}

	path_db := fmt.Sprintf("%s?_journal=WAL", mconfig.path_db)
	log.Println("Opening SQLite DB at ", path_db)
	db := db_init(path_db)
	defer func() {
		if err := db.Close(); err != nil {
			log.Println("warning: error when closing database: ", err)
		}
	}()

	metrics, err := config.parse_metrics()
	if err != nil {
		log.Fatal("config file reading failed, cannot proceed with measure: ", err)
	}
	if err := db_migrate(db, metrics); err != nil {
		log.Fatal("cannot proceed with measure: ", err)
	}

	log.Println("database retention period is ", mconfig.retention_time)

	ctx, cf := context.WithCancel(context.Background())

	ci := make(chan os.Signal, 1)
	signal.Notify(ci, os.Interrupt)
	go func() {
		for range ci {
			cf()
			fmt.Println("got SIGINT -- bailing")
		}
	}()

	ct := make(chan db_task)
	go db_writer(ctx, db, ct)
	go db_pruner(ctx, ct, metrics, mconfig.retention_time, mconfig.prune_db_period)
	run_metrics(ctx, db, mconfig.measure_period, mconfig.shell, metrics, ct)
}
