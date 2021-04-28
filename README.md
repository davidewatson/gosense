# gosense

`gosense` is a lightweight hardware monitor. It aims to deliver a small
statically linked binary (cf. 7.9 MB for go version go1.13 linux/amd64, 2.1 MB
stripped and compressed), which has lower runtime overhead and increased safety
when compared with alternatives.

Each monitor implements a cache with an asynchronous goroutine to periodically
update itself. A shared REST server services requests from all monitors. The
cache allows for low overhead constant time observations. This is important for
some monitors (e.g. hard disks) because updates may take many minutes.

# tl;dr

```bash
buck build //experimental/dwat/gosense
buck test //experimental/dwat/gosense/...
```

```bash
buck run //experimental/dwat/gosense &
curl --insecure localhost:8080/api/sys/sensors
fg
^C
```

# Implementation

Two example monitors can be found [here](./pkg/lmsensors/lmsensors.go) and
[here](./pkg/lmsensors/classic/classic.go).

- Creating a monitor is as simple as writing an `Update() ([]byte, error)`
function, starting a cache with it, and registering an HTTP handler. See
[main.go](./main.go) for an example.

- Update functions may block, though functions which timeout will panic.

- Update functions which fail to run successfully the first time will be
ignored.

- After an update returns an error the cache will not return stale data.
Instead it will return the error packaged in a JSON object until the next
successful update.

# Design

- Monitoring is a p0 capability. This means it must work no matter what else
has failed, because we will depend on it to get the system back into a healthy
state. Therefore there can be no external dependencies.

- Monitors should have minimal overhead so as not to disturb the actual
workload.

- Non-responsive monitors should panic instead of failing silently.

- Complex or unsafe monitoring should be done within a process to protect
other monitors sharing the same http server process.

- Consider using libraries or other API based data sources when available since
they are often lower overhead and higher fidelity.

# Security

TODO(dwat): Configure http server to load a certificate.

# Performance

## Benchmarks

```bash
go test -cpuprofile cpu.prof -memprofile mem.prof -benchmem -bench .
```

### lmsensors library

```bash
goos: linux
goarch: amd64
pkg: experimental/dwat/gosense/pkg/lmsensors
Benchmark/c.Get()-56         	790041016	         1.50 ns/op	       0 B/op	       0 allocs/op
Benchmark/c.UpdateWithTimeout()-56         	     289	   4102465 ns/o  921400 B/op	    1883 allocs/op
PASS
ok  	experimental/dwat/gosense/pkg/lmsensors	3.118s
```

### classic fork and parse `sensors` cli

Note that the cost of the `sensors` command is not measured (because it runs in
another process). TODO(dwat): measure.

```bash
goos: linux
goarch: amd64
pkg: experimental/dwat/gosense/pkg/lmsensors/classic
Benchmark/c.Get()-56         	774242150	         1.46 ns/op	       0 B/op	       0 allocs/op
Benchmark/c.UpdateWithTimeout()-56         	     192	   6261189 ns/o   75265 B/op	     247 allocs/op
PASS
ok  	experimental/dwat/gosense/pkg/lmsensors/classic	3.167s
```

## Profiling

```bash
go tool pprof -http="$(hostname):9090" ./gosense memprofile.out
open localhost:9090 &
```

## Even smaller binaries

Cf. https://blog.filippo.io/shrink-your-go-binaries-with-this-one-weird-trick/

```bash
go build -ldflags="-s -w"
upx ./gosense
ls -lh ./gosense
```

```bash
                       Ultimate Packer for eXecutables
                          Copyright (C) 1996 - 2018
UPX 3.95        Markus Oberhumer, Laszlo Molnar & John Reiser   Aug 26th 2018

File size         Ratio      Format      Name
--------------------   ------   -----------   -----------
6008832 ->   2158064   35.91%   linux/amd64   gosense

Packed 1 file.
-rwxr-xr-x. 1 dwat users 2.1M Sep 24 19:18 ./gosense
```

# References

## Linux Monitor Sensors (lmsensors)

GitHub: https://github.com/mdlayher/lmsensors
Wiki: https://wiki.archlinux.org/index.php/Lm_sensors
FAQ: https://hwmon.wiki.kernel.org/faq
