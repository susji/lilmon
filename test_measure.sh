#!/bin/sh

CONFIG=test_measure.ini
DB="$(grep -F path_db= $CONFIG | cut -d '=' -f2)"
WAITSECS=7

set -eu

if ! command -v sqlite3 > /dev/null; then
    echo "missing sqlite3"
    exit 1
fi

if ! make ; then
    echo "build failed"
    exit 2
fi

if [ ! -e lilmon ]; then
    echo "missing lilmon"
    exit 3
fi

if [ -z "$DB" ]; then
    echo "missing db config"
    exit 4
fi

echo db=$DB

clean_db() {
    rm -f $DB ${DB}-shm ${DB}-wal
}

clean_db
./lilmon measure -config-path "$CONFIG" &
LMPID=$!
echo LMPID=$LMPID

sleep $WAITSECS
echo killing $LMPID
kill $LMPID

val=$(sqlite3 "$DB" 'SELECT CAST(value AS INT) FROM lilmon_metric_test' | tail -1)
clean_db

if [ "$val" != "12" ]; then
    echo "value mismatch, got $val"
    exit 10
fi

echo "test ok"

