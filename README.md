# lilmon

## What is it?

lilmon is a small program for collecting and displaying numeric values as time
series graphs in a browser view. If you stretch the definition a bit, it is a
monitoring tool.

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

## How does it look like?

The graphs are drawn using `gonum.org/v1/plot`. Currently lilmon produces graphs
like this:

![screenshot of lilmon
UI](https://github.com/susji/lilmon/raw/main/lilmon.png "lilmon v0.x.y")

## Does lilmon do alerting?

No. Its intended purpose is to record numeric values and display them with a
bare bones UI. However, as everything is recorded into a SQLite database, a
different program can easily follow the metrics and do alerting based on that.

## Will lilmon have a configuration UI?

No.

## The graphs look terrible!

~~Yes. I'll probably make them less terrible in future.~~
Much better now, right?

## Will lilmon support monitoring more than one machine?

As all lilmon metrics are just columns in a SQLite table, they can be
transferred outside their host of origin with relative ease. It's just not
something I'm especially interested in.

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

Given above, you could then configure a metric like this:
```
[metrics]
metric=n_id_chars|Characters output by privileged id|y_min=0|doas /usr/bin/id|wc -c
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
  - `kibi` and `kilo`: Y values are rendered with unit prefixes in base-2 or base-10, respectively

`deriv` is useful if your metric is, for example, measuring transmitted or
received bytes for a network interface. By using `deriv`, the UI will then
display transfer rates (bytes/second) instead of bytes.

`kibi` and `kilo` will make larger values much more easier to read.

## Can you edit the browser UI?

Yes, just use [the example as basis](lilmon.template.example) and have at it.

## What is required to run lilmon?

**NOTE**: lilmon is currently very experimental software and it is not packaged
in any manner. Your usage experience will be mildly tedious. I will fix this in
near future.

These manual steps should be roughly enough to run lilmon experimentally in
whatever UNIX-like environment is supported by the Go toolchain. Please note
that the pure-Go SQLite library also sets some limitations for platform support.

Below we assume GNU/Linux, so please note any differences for *BSD. As mentioned
above, please do not run lilmon as root or any other unnecessarily privileged
user.

Below we assume that `root` is the privileged user and its primary group is
`root`. The group could be something else, too, like `wheel`. Please note that
the `#` prefix in the example commands does mean a root shell according to the
tradition.

1. Obtain a `lilmon` executable -- possibly `go build` is enough, see Go's
   cross-compiling instructions if you need to target different OS/arch

2. Install the binary on your target machine:

```
# install -m 0755 -o root -g root lilmon /usr/local/bin
```

3. Create `/etc/lilmon.ini` -- use [the example file](lilmon.ini.example) as
   basis and make sure it is writable only by privileged users. Probably this
   means it should be owned by `root:root` or something similar.

```
# install -T -m 0644 -o root -g root lilmon.ini.example /etc/lilmon.ini
```

4. Copy the browser UI's HTML template to a privileged directory:

```
# install -T -m 0644 -o root -g root lilmon.template.example /etc/lilmon.template
```

5. Create a new non-privileged system user and group for lilmon:

```
# adduser --disabled-login --system --no-create-home --group lilmon
```

6. Create a directory suitable for storing the lilmon database:

```
# mkdir /var/lilmon
# chown lilmon:lilmon /var/lilmon
```

7. Begin measuring as the `lilmon` user:

```
# sudo -u lilmon /usr/local/bin/lilmon measure
```

8. Begin serving the monitoring interface as the `lilmon` user. Please note that
   `lilmon serve` by default only listens on `localhost:15515`:

```
# sudo -u lilmon /usr/local/bin/lilmon serve -addr "${LISTEN_ADDR}:15515"
```

where `$LISTEN_ADDR` should be a suitable listening address or `0.0.0.0` if
you want lilmon to listen on all addresses.

9. Point your browser at the listener.

## What about TLS, rate limiting, authentication...?

I strongly recommend a reverse proxy for handling these things.

## Known limitations

- If a metric is disabled by removing it from the configuration file, its
  historical data will not be automatically pruned after the retention period

## TODO

- [ ] cache graphs
- [ ] support units for smart Y labels (eg. "bytes")
