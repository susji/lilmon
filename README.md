# lilmon

lilmon is a small self-contained program for collecting and displaying numeric
values as time series graphs in a browser view.

It has two modes of operation:
1. measure
2. serve

In the measure mode, lilmon records numeric for values for its configured
metrics. It inserts these values into a SQLite database.

In the serve mode, lilmon displays dynamically rendered time series graphs for
the configured metrics.

Each metric has a name, description, graphing options, and a command. Commands
is shell-expanded like `$SHELL -c ${command}` and lilmon expects to receive a
single number back.

lilmon is not intended to be an alerting platform. It exists just to draw some
simple graphs for UNIX-like environments.

# Mode of operation

Please note that lilmon executes the metrics commands as the user it is started
as. It does not do any kind of privilege separation. If you start `lilmon
measure` as root, your commands will be run as root. Similarly, if you start
`lilmon serve` as a root, the browser interface will also be run as root.

It is not necessary to run either of the modes as a privileged user. For the
measure mode, we suggest you to use `sudo`, `doas`, or something similar with
limited capabilities to obtain privileged metrics.

As an example, with `doas` you may permit the `_lilmon` user to run
`/usr/bin/id` without any arguments as `root` like this:

```doas
permit nopass _lilmon as root cmd /usr/bin/id args
```

# Configuration

See [the example file](lilmon.ini.example) for inspiration. Each definition of a
metric consists of the following four fields:

    <name>|<description>|<graph-options>|<raw-shell-command>

The shell command may contain `|` characters -- it will not affect configuration
parsing.

`<graph-options>` may contain the following `,` separated parameters:

  - `deriv`: The time series is numerically differentiated with respect to time
  - `y_min=<float64>`: Graph's minimum Y value
  - `y_max=<float64>`: Graph's maximum Y value

`deriv` is useful if your metric is, for example, measuring transmitted or
received bytes for a network interface. `deriv` will then display bytes/seconds
instead.

# Usage

These are the essential steps, which may also be automated.

1. Create a configuration file in some secure location. The default is
   `/etc/lilmon.ini`. Make sure it is writable **only** by the intended,
   privileged users.
2. Create a non-privileged daemon user and group such as `_lilmon:_lilmon`.
3. Determine a suitable location for the metrics database. The default is
   `/var/lilmon/lilmon.sqlite`. Make sure the directory exists. Again, do this
   with suitable filesystem permissions.
4. Run the `measure` mode as the `_lilmon` user with
```
$ lilmon measure
```
5. Run the `serve` mode as the `_lilmon` user with
```
$ lilmon serve -template-path ./static/serve.template
```
6. Point your browser at `http://localhost:15515`

If you wish use non-default parameters such as different location for the
database file, please consult `lilmon measure -h` and `lilmon serve -h`.

# Known limitations

- If a metric is disabled by removing it from the configuration file, its
  historical data will not be automatically pruned after the retention period

# TODO

- [ ] cache graphs
- [ ] slightly more responsive html
- [ ] render a bit more guiding ticks for graphs
- [ ] implement some logic to filter out outlier data points as a graph option for noisy data
- [ ] support units for smart Y labels (eg. "bytes")
- [x] make graph parametes configurable with the file
- [x] graph timestamp label based on range size
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
