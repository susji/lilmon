# lilmon

This is a little monitoring system.

# Known limitations

- If a metric is disabled by removing it from the configuration file, its
  historical data will not be automatically pruned after the retention period
- The graphs are static and ugly

# TODO

- [x] prune metric tables for data beyond retention time
- [x] index tables with timestamps
- [x] programmatically generate graphs
- [x] implement simple `serve` subcommand to display a monitoring interface
- [x] make running as `root` harder
- [x] configure metrics with a file
- [x] draw min & max labels for graph axis
- [ ] render HTML with proper templates
- [ ] cache graphs
- [ ] insert HTML for changing the graphed time range
