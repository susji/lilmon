# lilmon

`lilmon` is a small self-contained program for collecting and displaying numeric
values as time series graphs in a browser view.

It has two modes of operation:
1. `measure`
2. `serve`

In `measure` mode, `lilmon` records numeric for values for its configured
metrics.

In `serve` mode, `lilmon` displays dynamically rendered time series graphs for the configured metrics.

Each metric has a `name`, `description`, graphing `options` `command`. The
`command` is shell-expanded to obtain some kind of a numeric value.

# Mode of operation

Please note that `lilmon` executes the metrics commands as the user it is
started as. It does not do any kind of privilege separation. If you start
`lilmon measure` as `root`, your commands will be run as `root`. Similarly, if
you start `lilmon serve` as a `root`, the browser interface will also be run as
`root`.

Neither modes need a privileged user to be run successfully. For the `measure`
mode, you may want to employ `sudo` or `doas` to obtain privileged metrics.

# Configuration

See [the example file](lilmon.ini.example) for inspiration.

# Usage

These are the essential steps, which may also be automated.

1. Create a configuration file in some secure location. The default is
   `/etc/lilmon.ini`. Make sure it is writable only by the intended, privileged
   users.
2. Determine a suitable location for the metrics database. The default is
   `/var/lilmon/lilmon.sqlite`. Again, do this with suitable filesystem
   permissions.
3. Run the `measure` mode with
```
$ lilmon measure
```
4. Run the `serve` mode with
```
$ lilmon serve -template-path ./static/serve.template
```
5. Point your browser at `http://localhost:15515`

If you wish use non-default parameters such as different location for the
database file, please consult `lilmon measure -h` and `lilmon serve -h`.

# Known limitations

- If a metric is disabled by removing it from the configuration file, its
  historical data will not be automatically pruned after the retention period

# TODO

- [ ] cache graphs
- [ ] slightly more responsive html
- [ ] render a bit more guiding ticks for graphs
- [ ] make graph drawing things configurable after proper templating is done
- [ ] implement some logic to filter out outlier data points
- [ ] graph timestamp label based on range size
- [x] permit setting individual Y limits for graph rendering
- [x] make graph-binning relative to time range
- [x] render HTML with proper templates
- [x] support derivatives of metrics (like interface speed is a derivative of aggregate bytes)
- [x] prune metric tables for data beyond retention time
- [x] index tables with timestamps
- [x] programmatically generate graphs
- [x] implement simple `serve` subcommand to display a monitoring interface
- [x] make running as `root` harder
- [x] configure metrics with a file
- [x] draw min & max labels for graph axis
- [x] insert HTML for changing the graphed time range
