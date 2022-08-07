# Changelog

## 0.7.0 (2022-08-07)

### Changes

* Graphs include a grid now

### Additions

* Give possibility to render neater Y values with new graphing options `kibi`
  and `kilo` which render neater numbers in base-2 and base-10

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
