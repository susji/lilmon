# lilmon

## What is it?

lilmon is a small program for collecting numeric values on your UNIX-like system
and displaying them as time series in a browser view. If you stretch the
definition a bit, it is a minimalistic monitoring tool.

lilmon is currently very experimental software.

## Why does it exist?

I needed a small monitoring tool for my own use.

## What does it do?

lilmon has two modes of operation.

In the first mode, `measure`, it periodically executes commands for its metrics
and gathers their numeric values into a SQLite database.

In the second mode, `serve`, it displays the recorded values with a dynamic HTML
page.


## What does lilmon measure?

lilmon measures numeric values. To do this, lilmon is given a set of *metrics*.
Each metric has a

- name
- description
- graphing options
- command

lilmon uses raw shell-commands to obtain these numeric values for each specified
metric. Commands are shell-expanded like

    $ /bin/sh -c '<metric-command>'

and as a result, lilmon expects to receive a single value back via `stdout`. The
value is interpreted as a `float64`. Integers are also fine. Whitespace is
trimmed before any interpretation is attempted.

A minimalistic example of a metric command would then be `echo 123` which would
result in a static value of `123` on each measurement.

## How does it look like?

The graphs are drawn using `gonum.org/v1/plot`. Currently lilmon produces graphs
like this:

![screenshot of lilmon
UI](https://github.com/susji/lilmon/raw/main/lilmon.png "lilmon v0.x.y")

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
  - `kibi` and `kilo`: Y values are rendered with unit prefixes in base-2 or base-10, respectively

`deriv` is useful if your metric is, for example, measuring transmitted or
received bytes for a network interface. By using `deriv`, the UI will then
display transfer rates (bytes/second) instead of bytes.

`kibi` and `kilo` will make larger values much more easier to read.

## Show me some example metrics!

These are some metrics I use. They may fail in cases I have not thought about.
There are many ways to obtain similar results. The primary reason for including
these here is to give you inspiration on how to use lilmon.

As is stressed elsewhere in this README, to be safer, do not run any of these as
a privileged user.

### TX & RX speed of some interface

Use the correct interface name in place of `if-name`. If you need to measure
than one interface, define more similar metrics with different metric name and
`if-name`. Silly but simple!

Also, note the `deriv` in the graphing options, which means that the raw byte
counts are numerically differentiated when the graph is drawn. The result is
then a decent approximation of TX & RX speed.

#### Linux

```
metric=bytes_wifi_rx|Wifi RX|y_min=0,deriv,kilo|cat /proc/net/dev|fgrep if-name|awk '{print $2}'
metric=bytes_wifi_tx|Wifi TX|y_min=0,deriv,kilo|cat /proc/net/dev|fgrep if-name|awk '{print $10}'
```

#### OpenBSD

```
metric=bytes_wired_rx|Wired RX|deriv,y_min=0,kilo|netstat -n -i -b|fgrep if-name|fgrep Link|awk '{print $5}'
metric=bytes_wired_tx|Wired TX|deriv,y_min=0,kilo|netstat -n -i -b|fgrep if-name|fgrep Link|awk '{print $6}'
```

### Temperature sensor

#### Linux

The example here makes use of `jq` to search the JSON dump produced by `sensors
-j`. See what `sensors -j` displays for you and accomodate the `jq` filter. You
may of course produce the same sensor value by just parsing and filtering the
regular text dump.

```
metric=temp_cpu|CPU temperature|y_min=30,y_max=90|sensors -j|jq '.["dev::temp1::temp1_input"]'
```

#### OpenBSD

Look at the output of `sysctl hw.sensors` and figure out the the exact path for
your device. If it has something other than a raw float value, filter the rest
out.

```
metric=cpu_temp|CPU temperature|y_min=40,y_max=90|sysctl hw.sensors.km0.temp0|cut -d '=' -f2|cut -d ' ' -f 1
```

### Ping round-trip time for a well-known target

#### Linux

Note that in this example we use the `-w 10` option to define a hard deadline of
10 seconds. This is not fully portable, so see your `man 8 ping` for more
details. Something like the `timeout` command is available on many platforms,
and it works well for making sure programs time out.

```
metric=ping_google|PING Google|y_min=0,kilo|ping -q -w 10 -c 2 8.8.8.8|tail -1|cut -d'=' -f2|cut -d '/' -f2
```

### System load (1 min)

#### OpenBSD

```
metric=load_1|1 minute CPU LOAD|y_min=0|uptime|grep -E -o 'averages: [\.0-9]+'|cut -d ' ' -f2
```

### Free memory

#### OpenBSD

For an example's sake, we go through some trouble to dig out some bytes. Perhaps
we are lucky and `top` always prints megabytes?

```
metric=free_mem|Free memory|y_min=0,kilo|echo $((1024 * 1024 * $(top -b|egrep -o 'Free: [0-9]+'|cut -d ' ' -f2)))
```

### Wi-Fi clients (stations) connected to hostapd

Note that this invocation probably requires privileged execution. See the
discussion regarding `sudo` and `doas`. This command also assumes that there is
only one station or that the station of interest has index `0`.

```
metric=n_sta|hostapd clients|y_min=0|sudo /usr/sbin/hostapd_cli status|fgrep 'num_sta[0]'|cut -d '=' -f 2
```

## Does lilmon do alerting?

No. Its intended purpose is to record numeric values and display them with a
bare bones UI. However, as everything is recorded into a SQLite database, a
different program can easily follow the metrics and do alerting based on that.
For details, see the next question.

## How do I access the values lilmon has gathered?

First make sure you have `sqlite3` installed. Then you can do something like
the following to get the 10 latest measurements for metric `METRIC_NAME`.

    $ sqlite3 'file:/var/lilmon/db/lilmon.sqlite?mode=ro' \
          'SELECT * FROM lilmon_metric_METRIC_NAME ORDER BY timestamp DESC LIMIT 10'

Note the `mode=ro` part for read-only.

## Will lilmon have a configuration UI?

No.

## The graphs look terrible!

~~Yes. I'll probably make them less terrible in future.~~
Much better now, right?

## `lilmon serve` refuses to start and says it cannot open the database!

Are you running `serve` without a pre-existing database? In that case, you will
have to start `lilmon measure` first because it will create the missing
database.

## How to proceed after changing the metrics in the configuration file?

Restart both processes but do restart `lilmon measure` first. It is responsible
for creating new database tables and their indexes for new or renamed metrics.

## Will lilmon support monitoring more than one machine?

As all lilmon metrics are just columns in a SQLite table, they can be
transferred outside their host of origin with relative ease. It's just not
something I'm especially interested in.

## How to run measurement commands which require privileged execution?

It is **not** required or recommended to run lilmon as a privileged user.

For the measure mode, it is wiser to make use of `sudo`, `doas`, or something
similar with limited capabilities to obtain privileged metrics. You would then
run these exactly as the non-privileged ones, except through `doas` like here:

```
[metrics]
metric=n_id_chars|Characters output by privileged id|y_min=0|doas /usr/bin/id|wc -c
```

### doas

With `doas` you may permit the `lilmon` user to run `/usr/bin/id` without any
arguments as `root` like this:

```doas
permit nopass lilmon as root cmd /usr/bin/id args
```

Do note that the lone keyword `args` suffixed to the command means only
execution without any arguments.

### sudo

`sudo` also permits specifying that the command can only be run without any
arguments, however as `sudo` also supports very rich logic for specifying how
and what commands can be run, please exercise caution, stick to simple
definitions, and give the relevant manuals a careful read. In this example, the
user `lilmon` is permitted to run `/usr/bin/id` without any arguments as `root`
and without a password on all hosts where the configuration file is active:

```sudo
lilmon ALL=(root) NOPASSWD: /usr/bin/id ""
```

Do note that if you do not specify the `""` suffix, then the above invocation
would permit `lilmon` to execute `/usr/bin/id` with arbitrary arguments.

## Can you edit the browser UI?

Yes, just use [the example as basis](lilmon.template.example) and have at it.

## What is required to run lilmon?

**NOTE**: lilmon is currently **very experimental software** and it is not yet
packaged in any reasonable manner. Your usage experience will be mildly tedious.
Before trying to perform an install with the attached `Makefile`, convince
yourself that it is doing the right thing. At this stage, performing a manual
install may be a better idea.

The installation is for the most part condensed into `make install`, but the
creation of the non-privileged user is platform-dependent. We also must give
that user a chance to write its database in the directory. For GNU/Linux it
looks like this

```
# make install
# adduser --disabled-login --system --no-create-home --group lilmon
# chown lilmon:lilmon /var/lilmon/db
# sudo -u lilmon /usr/local/bin/lilmon measure
# sudo -u lilmon /usr/local/bin/lilmon serve
```

When you are starting lilmon fresh without a pre-existing database, the first
run of `lilmon measure` will create it. As `lilmon serve` opens the database in
a read-only mode, it cannot initialize the database. Thus make sure have
successfully ran `measure` at least once before running `serve`.

Also note that by default `lilmon serve` listens only on localhost. You may want
to set the listening adress to something else such as a suitable interface's IP.
If you want it to listen on all interfaces, use `0.0.0.0:15515` but please do
not expose the lilmon browser view to any untrusted networks. As suggested
below, you may in any case wish to provide the actual access via a suitable
reverse proxy.

## Do I need timeouts for my commands?

It does not hurt, but lilmon tries to cancel measurement commands which take
`$TOO_LONG` to complete. See `metrics.go` for the details.

## What about TLS, rate limiting, authentication...?

I strongly recommend a reverse proxy for handling these things.

## Known limitations

- If a metric is disabled by removing it from the configuration file, its
  historical data will not be automatically pruned after the retention period

## TODO

- [ ] cache graphs
- [ ] support units for smart Y labels (eg. "bytes")
