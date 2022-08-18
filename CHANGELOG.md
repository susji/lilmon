# Changelog

## 0.12.0 (2022-08-18)

### Additions

* `measure` now uses `pledge`

### Changes

* `serve` should, just in case, get file-creation access to DB directory

### Fixes

* README is again better

## 0.11.1 (2022-08-17)

### Fixes

* `Makefile` is now aware of protect
* `pledge` & `unveil`: SQLite3 needs `/tmp` access

## 0.11.0 (2022-08-17)

### Additions

* Introduce some concept of dropping capabilities depending on platform
  * Currently this only means OpenBSD and `pledge` and `unveil`
  * First it will affect the `serve` mode
  * This feature need some iteration to be robust and effective

### Fixes

* README improvements
* `test_measure.sh` is slightly cleaner

## 0.10.2 (2022-08-15)

### Fixes

* `Stat` template before opening and panicking

### Additions

* `test_measure.sh` for running a short system test for measuring

## 0.10.1 (2022-08-13)

### Fixes

* `lilmon serve` will now refuse to start if it cannot `Stat` the database file
* `Makefile` is now `.PHONY` correct and we are explicit about its GNU style

### Changes

* README improvements

## 0.10.0 (2022-08-12)

### Additions

* Add `Makefile` to document installation and building

### Changes

* Change default configuration directory to `/etc/lilmon` instead of `/etc`
* Change default database directory to `/var/lilmon/db` instead of `/var/lilmon`
* Update README regarding installation

## 0.9.2 (2022-08-12)

### Fixes

* Unbreak metric table & index creation

### Changes

* Test initialization against a file-backed database
* Form fuller URI paths when opening the databaseb
* README is again more sensible

## 0.9.1 (2022-08-11)

### Changes

* Use `interface{}` in the tests instead of `any` for backwards compatibility
* Declare Go 1.17 instead of 1.18 as the module's *minimum-go-version*
* Refactor configuration handling for easier testing

### Fixes

* README: Minor improvements
* Minor naming improvements in code

## 0.9.0 (2022-08-10)

### Changes

* All configuration is now read from the configuration file

### Additions

* `lilmon.ini.example` has now gained several new key-value pairs

## 0.8.0 (2022-08-09)

### Changes

* Begin using `github.com/mattn/go-sqlite3` for SQLite

### Additions

* README: Included some realistic metric examples

## 0.7.2 (2022-08-08)

### Fixes

* Once more: Values should now be more readable

## 0.7.1 (2022-08-08)

### Fixes

* Values rendered with `kibi` and `kilo` should now make more sense

## 0.7.0 (2022-08-07)

### Changes

* Graphs include a grid now

### Additions

* Give possibility to render neater Y values with new graphing options `kibi`
  and `kilo` which render in units of base-2 and base-10, respectively

## 0.6.0 (2022-08-07)

### Changes

* When graphing, use more suitable timestamp formats
* Try to render timestamps using the local timezone
* As before, use hardcoded constants for graph colors
* Make graph format configurable and default to SVG
* Make `max_bins` configurable
* Update the example screenshot

## 0.5.0 (2022-08-06)

### Changes

* Start using `gonum.org/v1/plot` to draw the graphs
* README is again clearer
* Update the example screenshot

## 0.4.1 (2022-08-06)

### Fixes

* Disallow `-` in metric names as those will not work with our approach to
  SQLite queries

### Changes

* Bump `modernc.org/sqlite` to v1.18.0

## 0.4.0 (2022-08-06)

### Changes

* Start using `modernc.org/sqlite` for SQLite

## 0.3.0 (2022-08-05)

### Changes

* Implement more graceful handling of graph-drawing attempts without any bins
* Less error-prone handling for conjuring our database queries
* Default to using WAL with SQLite to avoid locking issues with readers vs. writers
* README is again better

## 0.2.0 (2022-07-31)

### Changes

* For consistency, renamed the example HTML template to
  `lilmon.template.example`
* Use adaptive timestamp format for HTML template
* Reworked the README
* Remove the unused `period` parameter of the measurement mode

### Additions

* Catch unrecognized configuration items
* Add a few example time ranges to the HTML template to display
  * the 24h before the past 24h
  * the week before the past week

## 0.1.0 (2022-07-30)

* First version of the code that has enough features for decent use
