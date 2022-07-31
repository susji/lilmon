# lilmon

## What is it?

lilmon is a small program for collecting and displaying numeric values as time
series graphs in a browser view. If you stretch the definition a bit, it is a
monitoring tool.

lilmon is currently very experimental software.

## Why does it exist?

I needed a small monitoring tool for my own use.

## What does it do?

lilmon has two modes of operation. In the first mode, `measure`, it executes
commands for its metrics and gathers their numeric values into a SQLite
database. In the second mode, `serve`, it displays the configured metrics as
ascetic time series graphs as a browser interface.

## What does lilmon measure?

lilmon measures numbers. To do this, lilmon is given a set of *metrics*. Each
metric has a

- name
- description
- graphing options
- command.

lilmon uses raw shell-commands to obtain these numeric values for each specified
metric. Commands are shell-expanded like

    $ /bin/sh -c '<metric-command>'

and lilmon expects to receive a single number back. The number is interpreted as
a `float64` so integers are also fine. A minimalistic example of a metric
command would then be `echo 123` which would result in a static value of `123`
on each measurement.

## Does lilmon do alerting?

No. It's intended purpose is to record numeric values and display them with a
bare bones UI. However, as everything is recorded into a SQLite database, a
different program can easily follow the metrics and do alerting based on that.

## Warning

Please note that lilmon executes the metrics commands as the user it is started
as. It does not do any kind of privilege separation. If you start `lilmon
measure` as root, your commands will be run as root. Similarly, if you start
`lilmon serve` as a root, the browser interface will also be run as root.

It is not necessary to run either of the modes as a privileged user. For the
measure mode, we suggest you to use `sudo`, `doas`, or something similar with
limited capabilities to obtain privileged metrics.

As an example, with `doas` you may permit the `lilmon` user to run
`/usr/bin/id` without any arguments as `root` like this:

```doas
permit nopass lilmon as root cmd /usr/bin/id args
```

## How do you configure lilmon?

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
received bytes for a network interface. By using `deriv`, the UI will then
display transfer rates (bytes/second) instead of bytes.

## Can you edit the browser UI?

Yes, just use [the example as basis](lilmon.template.example) and have at it.

## What is required to run lilmon?

**NOTE**: lilmon is currently very experimental software and it is not packaged
in any manner. Your usage experience will be mildly tedious. I will fix this in
near future.

These manual steps should be roughly enough to run lilmon experimentally in
whatever UNIX-like environment is supported by the Go toolchain. Below we assume
GNU/Linux, so please note any differences for *BSD. As mentioned above, please
do not run lilmon as root or any other unnecessarily privileged user.

1. Obtain a `lilmon` executable -- possibly `go build` is enough, see Go's
   cross-compiling instructions if you need to target different OS/arch
2. Install the binary on your target machine:
```
# install lilmon /usr/local/bin/lilmon
```
3. Create `/etc/lilmon.ini` -- use [the example file](lilmon.ini.example) as
   basis and make sure it is writable only by privileged users. Probably this
   means it should be owned by `root:root` or something similar.
```
# install -m 0644 -o root -g root lilmon.ini.example /etc/lilmon.ini
```

4. Copy the browser UI's HTML template to lilmon's directory:
```
# install -m 0644 -o root -g root lilmon.template.example /etc/lilmon.template
```
5. Create a new non-privileged system user and group for lilmon:
```
# adduser --disabled-login --system --no-create-home --group lilmon
```
6. Create a directory suitable for storing the lilmon database and HTML
   Otemplate:
```
# mkdir /var/lilmon
# chown lilmon:lilmon /var/lilmon
```
7. Begin measuring as the `lilmon` user:
```
# sudo -u lilmon /usr/local/bin/lilmon measure
```
8. Begin serving the monitoring interface as the `lilmon` user. Please note that
   `lilmon serve` by default only listens on `localhost:15515`:`
```
# sudo -u lilmon /usr/local/bin/lilmon serve -addr "${LISTEN_ADDR}:15515""
```
9. Point your browser at `http://${LISTEN_ADDR}:15515`


# Known limitations

- If a metric is disabled by removing it from the configuration file, its
  historical data will not be automatically pruned after the retention period

# TODO

- [ ] cache graphs
- [ ] slightly more responsive html
- [ ] render a bit more guiding ticks for graphs
- [ ] implement some logic to filter out outlier data points as a graph option for noisy data
- [ ] support units for smart Y labels (eg. "bytes")
