# Changelog

## 0.17.1 (2024-04-01)

### Fixes

* When building a release, try less ambitious targets
  * This is due to our dependency CGO & SQLite

## 0.17.0 (2024-04-01)

### Changes

* Use GitHub release automation with GoReleaser
  * Pin Go to v1.22.1
  * Build for amd64, arm64, and arm on Linux, OpenBSD, and FreeBSD
* Lots of hopefully meaningless dependency bumps

## 0.16.0 (2022-09-11)

### Fixes

* `Makefile` fully fixed to prefer `/etc/lilmon/`
* More robust logic around graph binning

### Changes

* Example template is again more responsive
* Draw scatter bars instead of plain glyphs
* Double the default value of `downsampling_scale` to 4
* More aggressive logging around downsampling

### Additions

* Graph colors are now configurable!

## 0.15.0 (2022-09-03)

### Changes

* Update `go-sqlite3` from v1.14.12 to v1.14.15
* Layout in the example template is now easier to the eyes and more responsive

### Additions

* Added `no_ds` option also as an URL query parameter

## 0.14.1 (2022-08-21)

### Fixes

* Enforce proper values for `downscaling_scale` & clarify its usage
* Enforce proper value for `measure_period`

## 0.14.0 (2022-08-21)

### Changes

* When graphing, long queries are now automatically downsampled
  * The goal is to make graph generation much faster than it was
  * The behavior may be turned off for individual metrics with the `no_ds`
    option
  * Downscaling effect may be decreased by increasing the global option
    `downscaling_scale`
  * This functionality will probably be at least finetuned after some testing
* As a side-effect of the downscaling change, `measure_period` now moved in the
  configuration file's global section
  * Older configuration files will produce a syntax error as `measure_period`
    used to live in the `[measure]` section
* Miscellaneous stuff has moved under `misc`

### Additions

* `misc/db_fill.py` for generating some test data

### Fixes

* Removed garbage from graph generation log messages

## 0.13.1 (2022-08-20)

### Fixes

* Minor logging issue with `measure`

## 0.13.0 (2022-08-20)

### Fixes

* If `serve` fails on HTTP listen, it no longer fails silently

### Changes

* `serve` and its database queries experienced some micro-optimization
* README is again better
* The example template gained a few additional time ranges
* Autorefresh default made slower (2 min)
* Measure less often by default
* Draw more bins by default

## 0.12.1 (2022-08-18)

### Fixes

* `measure` probably needs more `pledge` access for SQLite

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
