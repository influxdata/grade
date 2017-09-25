# grade

grade ingests Go benchmark data into InfluxDB so that you can track performance over time.

## Installation

To download and install the `grade` executable into your `$GOPATH/bin`:

```sh
go get github.com/influxdata/grade/cmd/grade
```

## Usage

### Initial database configuration

The data from Go benchmarks tends to be very time-sparse (up to perhaps dozens of commits per day),
so we recommend creating your database with an infinite retention and a large shard duration.
Issue this command to your InfluxDB instance:

```
CREATE DATABASE benchmarks WITH DURATION INF SHARD DURATION 90d
```

### Running the command

Although you can pipe the output of `go test` directly into `grade`,
for now we recommend placing the output of `go test` in a file first so that if something goes wrong,
you don't have to wait again to run all the benchmarks.

For example, to run all the benchmarks in your current Go project:

```sh
go test -run=^$ -bench=. -benchmem ./... > bench.txt
```

Then, assuming you are in the directory of your Go project and 
git has checked out the same commit corresponding with the tests that have run,
this is the bare set of options to load the benchmark results into InfluxDB via `grade`:

```sh
grade \
  -hardwareid="my dev machine" \
  -goversion="$(go version | cut -d' ' -f3-)" \
  -revision="$(git log -1 --format=%H)" \
  -timestamp="$(git log -1 --format=%ct)" \
  < bench.txt
```

Notes on this style of invocation:

* `-influxurl` is not provided but defaults to `http://localhost:8086`.
Basic auth credentials can be embedded in the URL if needed.
HTTPS is supported; supply `-insecure` if you need to skip SSL verification.
* `-database` is not provided but defaults to `benchmarks`.
* `-measurement` is not provided but defaults to `go`.
* The hardware ID is a string that you specify to identify the hardware on which the benchmarks were run.
* The Go version subcommand will produce a string like `go1.6.2 darwin/amd64`, but you can use any string you'd like.
* The revision subcommand is the full SHA of the commit, but feel free to use a git tag name or any other string.
* The timestamp is a Unix epoch timestamp in seconds.
The above subcommand produces the Unix timestamp for the committer of the most recent commit.
This assumes that the commits whose benchmarks are being run, all are ascending in time;
git does not enforce that commits' timestamps are ascending, so if this assumption is broken,
your data may look strange when you visualize it.

## Schema

For each benchmark result from a run of `go test -bench`:

* Tags:
	* `goversion` is the same string as passed in to the `-goversion` flag.
	* `hwid` is the same string as passed in to the `-hardwareid` flag.
	* `name` is the name of the benchmark function, stripped of the `Benchmark` prefix.
	* `ncpu` is the number of CPUs used to run the benchmark.
	* `pkg` is the name of Go package containing the benchmark, e.g. `github.com/influxdata/influxdb/services/httpd`.
* Fields:
	* `alloced_bytes_per_op` is the allocated bytes per iteration of the benchmark.
	* `allocs_per_op` is how many allocations occurred per iteration of the benchmark.
	* `mb_per_s` is how many megabytes processed per second when running the benchmark.
	* `n` is the number of iterations in the benchmark.
	* `ns_per_op` is the number of wall nanoseconds taken per iteration of the benchmark.
	* `revision` is the git revision specified in the `-revision` flag.
	This was chosen to be a field so that the information is quickly available but not at the cost of a growing series cardinality per benchmark run.
