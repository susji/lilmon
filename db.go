package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

func db_path_measure(filepath string) string {
	return fmt.Sprintf("file:%s?_journal=WAL", filepath)
}

func db_path_serve(filepath string) string {
	return fmt.Sprintf("file:%s?mode=ro", filepath)
}

func db_table_name_get(metric *metric) string {
	if !is_metric_name_valid(metric) {
		panic(fmt.Sprintf("invalid metric name: %#v", metric))
	}
	return fmt.Sprintf("lilmon_metric_%s", metric.name)
}

func get_oversampling(bins, scale int, measure_period time.Duration, time_start, time_end time.Time) int {
	// `osr` means oversampling ratio. It means how many actual measurement
	// samples we expect to find in one bin. We then use it to estimate what
	// fraction of samples we can drop off to achieve roughly the same
	// average result.
	//
	// To have some adjustability, we provide `scale` as a factor. It's
	// essentially making our measurement period seems greater. This means
	// increasing the scale will decrease the amount of downsampling.
	//
	// The reason for this effort is that this approach will keep the DB
	// query times sensible and we expect that most of our data is "smooth"
	// enough to survive this kind of downsampling.
	//
	// We also assume here silently that most of the data has been gathered
	// using roughly the same measurement period.
	//
	dt := time_end.Sub(time_start).Seconds()
	binwidth := dt / float64(bins)
	osr := int(binwidth / (measure_period.Seconds() * float64(scale)))
	log.Println("dt=", dt, "bins=", bins, "binwidth=", binwidth, "osr=", osr)
	return osr
}

func db_datapoints_get(db *sql.DB, metric *metric, scale, bins int, measure_period time.Duration, time_start, time_end time.Time) (
	[]datapoint, error) {

	template_select_values := `
SELECT timestamp, value FROM %s
    WHERE
        timestamp BETWEEN
            DATETIME(%d, 'unixepoch')
            AND DATETIME(%d, 'unixepoch')
        %s
    ORDER BY timestamp ASC`

	// ds means downsampling
	ds := ""
	if !metric.options.no_downsample {
		osr := get_oversampling(bins, scale, measure_period, time_start, time_end)
		if osr >= 2 {
			// We take the scaled multiplicative inverse of OSR and use it
			// to drop random samples.
			drop_abs := 10000
			drop_rel := float64(drop_abs) / float64(osr)
			if drop_rel < 1 {
				drop_rel = 1.0
			}
			ds = fmt.Sprintf(
				`AND (ABS(RANDOM()) %% %d) < %d`,
				drop_abs,
				int(drop_rel))
		}
	}

	q := fmt.Sprintf(
		template_select_values,
		db_table_name_get(metric),
		time_start.Unix(),
		time_end.Unix(),
		ds)

	rows, err := db.Query(q)
	if err != nil {
		log.Println("graph_generate: unable to select rows: ", err)
		return nil, err
	}
	defer rows.Close()
	dps := []datapoint{}
	for rows.Next() {
		var ts time.Time
		var value float64

		if err := rows.Scan(&ts, &value); err != nil {
			log.Println("graph_generate: row scan failed: ", err)
			break
		}
		dps = append(dps, datapoint{ts: ts, value: value})
	}
	return dps, nil
}

func db_init(db_path string) *sql.DB {
	db, err := sql.Open("sqlite3", db_path)
	if err != nil {
		log.Fatalf("cannot open database: %v\n", err)
	}

	var db_version string
	if err := db.QueryRow("SELECT sqlite_version()").Scan(&db_version); err != nil {
		log.Println("warning: unable to get sqlite version: ", err)
	} else {
		log.Println("database version: ", db_version)
	}
	return db
}

func db_migrate(db *sql.DB, metrics []*metric) error {
	template_table := `
CREATE TABLE IF NOT EXISTS %s (
    id INTEGER PRIMARY KEY,
    value DOUBLE PRECISION,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP);
CREATE INDEX IF NOT EXISTS index_%s ON %s (value, timestamp);`
	in_err := false
	for n, m := range metrics {
		log.Printf(
			"Maybe creating tables and indices for metric %d/%d: %s (%s)\n",
			n+1, len(metrics), m.name, m.description)

		tn := db_table_name_get(m)
		q := fmt.Sprintf(template_table, tn, tn, tn)
		if _, err := db.Exec(q); err != nil {
			log.Printf("failed to create table/index for metric %s: %v ", m.name, err)
			in_err = true
		}
	}
	if in_err {
		return errors.New("database migration encountered errors")
	}
	return nil
}

func db_writer(ctx context.Context, db *sql.DB, tasks <-chan db_task) {
	template_insert := `INSERT INTO %s (value) VALUES (?)`
	template_prune := `DELETE FROM %s WHERE timestamp < DATETIME('now', '-%d seconds')`
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-tasks:
			switch task.kind {
			case DB_TASK_INSERT:
				metric := task.insert_measurement.metric
				value := task.insert_measurement.value
				_, err := db.ExecContext(
					ctx,
					fmt.Sprintf(template_insert, db_table_name_get(metric)),
					value)
				if err != nil {
					log.Printf(
						"metric insert failed for %s with value %f: %v\n",
						metric.name, value, err)
				}
			case DB_TASK_PRUNE_TABLE:
				metric := task.prune_metric
				retention_period := task.prune_retention_period

				log.Printf(
					"Pruning metric %s for older than %s entries.\n",
					metric.name, retention_period)
				q := fmt.Sprintf(
					template_prune,
					db_table_name_get(metric),
					int64(retention_period/time.Second))
				_, err := db.ExecContext(ctx, q)
				if err != nil {
					log.Println("Pruning failed: ", err)
				}
			default:
				panic(fmt.Sprintf("This is a bug: db_task.kind == %d", task.kind))
			}
		}
	}
}

func db_pruner(ctx context.Context, tasks chan<- db_task, metrics []*metric,
	retention_period, prune_period time.Duration) {

	log.Println("Entering pruning loop with period of ", prune_period)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(prune_period):
			for _, m := range metrics {
				tasks <- db_task{
					kind:                   DB_TASK_PRUNE_TABLE,
					prune_metric:           m,
					prune_retention_period: retention_period,
				}
			}
		}
	}
}
