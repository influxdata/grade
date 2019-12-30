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
  -branch="$(git rev-parse --abbrev-ref HEAD)" \
  < bench.txt
```

Notes on this style of invocation:

* `-influxurl` is not provided but defaults to `http://localhost:8086`.
Basic auth credentials can be embedded in the URL if needed.
HTTPS is supported; supply `-insecure` if you need to skip SSL verification.
If you set it to an empty string, `grade` will print line protocol to stdout.
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
* The branch subcommand is the name of the current branch. The `-branch` flag is optional.


## Schema

For each benchmark result from a run of `go test -bench`:

* Tags:
	* `goversion` is the same string as passed in to the `-goversion` flag.
	* `hwid` is the same string as passed in to the `-hardwareid` flag.
	* `name` is the name of the benchmark function, stripped of the `Benchmark` prefix.
	* `pkg` is the name of Go package containing the benchmark, e.g. `github.com/influxdata/influxdb/services/httpd`.
	* `procs` is the number of CPUs used to run the benchmark. This is a tag because you are more likely to group by `procs` rather than chart them over time.
	* `branch` is the same string as passed in to the `-branch` flag.
	Since the `-branch` flag is optional and can be omited, the tag will be present only if the flag is set.
* Fields:
	* `alloced_bytes_per_op` is the allocated bytes per iteration of the benchmark.
	* `allocs_per_op` is how many allocations occurred per iteration of the benchmark.
	* `mb_per_s` is how many megabytes processed per second when running the benchmark.
	* `n` is the number of iterations in the benchmark.
	* `ns_per_op` is the number of wall nanoseconds taken per iteration of the benchmark.
	* `revision` is the git revision specified in the `-revision` flag.
	This was chosen to be a field so that the information is quickly available but not at the cost of a growing series cardinality per benchmark run.

## Sample

For a benchmark like this:

```
PASS
BenchmarkMarshal-2                  	  500000	      2901 ns/op	     560 B/op	      13 allocs/op
BenchmarkParsePointNoTags-2         	 2000000	       733 ns/op	  31.36 MB/s	     208 B/op	       4 allocs/op
BenchmarkParsePointWithPrecisionN-2 	 2000000	       627 ns/op	  36.68 MB/s	     208 B/op	       4 allocs/op
BenchmarkParsePointWithPrecisionU-2 	 2000000	       636 ns/op	  36.15 MB/s	     208 B/op	       4 allocs/op
BenchmarkParsePointsTagsSorted2-2   	 2000000	       947 ns/op	  53.85 MB/s	     240 B/op	       4 allocs/op
BenchmarkParsePointsTagsSorted5-2   	 1000000	      1189 ns/op	  69.75 MB/s	     272 B/op	       4 allocs/op
BenchmarkParsePointsTagsSorted10-2  	 1000000	      1624 ns/op	  88.05 MB/s	     320 B/op	       4 allocs/op
BenchmarkParsePointsTagsUnSorted2-2 	 1000000	      1167 ns/op	  43.69 MB/s	     272 B/op	       5 allocs/op
BenchmarkParsePointsTagsUnSorted5-2 	 1000000	      1627 ns/op	  50.99 MB/s	     336 B/op	       5 allocs/op
BenchmarkParsePointsTagsUnSorted10-2	  500000	      2733 ns/op	  52.32 MB/s	     448 B/op	       5 allocs/op
BenchmarkParseKey-2                 	 1000000	      2361 ns/op	    1030 B/op	      24 allocs/op
ok  	github.com/influxdata/influxdb/models	19.809s
```

Which is passed to `grade` like this:

```
grade \
  -influxurl '' \
  -goversion "$(go version | cut -d' ' -f3-)" \
  -hardwareid c4.large \
  -revision v1.0.2 \
  -timestamp "$(cd $GOPATH/src/github.com/influxdata/influxdb && git log v1.0.2 -1 --format=%ct)" \
  -branch="$(git rev-parse --abbrev-ref HEAD)" \
  < models-1.0.2.txt
```

You will see output like:
```
go,branch=master,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=Marshal,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=560i,allocs_per_op=13i,n=500000i,ns_per_op=2901,revision="v1.0.2" 1475695157000000000
go,branch=master,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=ParsePointNoTags,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=208i,allocs_per_op=4i,mb_per_s=31.36,n=2000000i,ns_per_op=733,revision="v1.0.2" 1475695157000000000
go,branch=master,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=ParsePointWithPrecisionN,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=208i,allocs_per_op=4i,mb_per_s=36.68,n=2000000i,ns_per_op=627,revision="v1.0.2" 1475695157000000000
go,branch=master,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=ParsePointWithPrecisionU,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=208i,allocs_per_op=4i,mb_per_s=36.15,n=2000000i,ns_per_op=636,revision="v1.0.2" 1475695157000000000
go,branch=master,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=ParsePointsTagsSorted2,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=240i,allocs_per_op=4i,mb_per_s=53.85,n=2000000i,ns_per_op=947,revision="v1.0.2" 1475695157000000000
go,branch=master,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=ParsePointsTagsSorted5,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=272i,allocs_per_op=4i,mb_per_s=69.75,n=1000000i,ns_per_op=1189,revision="v1.0.2" 1475695157000000000
go,branch=master,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=ParsePointsTagsSorted10,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=320i,allocs_per_op=4i,mb_per_s=88.05,n=1000000i,ns_per_op=1624,revision="v1.0.2" 1475695157000000000
go,branch=master,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=ParsePointsTagsUnSorted2,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=272i,allocs_per_op=5i,mb_per_s=43.69,n=1000000i,ns_per_op=1167,revision="v1.0.2" 1475695157000000000
go,branch=master,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=ParsePointsTagsUnSorted5,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=336i,allocs_per_op=5i,mb_per_s=50.99,n=1000000i,ns_per_op=1627,revision="v1.0.2" 1475695157000000000
go,branch=master,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=ParsePointsTagsUnSorted10,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=448i,allocs_per_op=5i,mb_per_s=52.32,n=500000i,ns_per_op=2733,revision="v1.0.2" 1475695157000000000
go,branch=master,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=ParseKey,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=1030i,allocs_per_op=24i,n=1000000i,ns_per_op=2361,revision="v1.0.2" 1475695157000000000
```


Or without the `-branch` flag:
```
grade \
  -influxurl '' \
  -goversion "$(go version | cut -d' ' -f3-)" \
  -hardwareid c4.large \
  -revision v1.0.2 \
  -timestamp "$(cd $GOPATH/src/github.com/influxdata/influxdb && git log v1.0.2 -1 --format=%ct)" \
  < models-1.0.2.txt
```

You will see output like:
```
go,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=Marshal,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=560i,allocs_per_op=13i,n=500000i,ns_per_op=2901,revision="v1.0.2" 1475695157000000000
go,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=ParsePointNoTags,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=208i,allocs_per_op=4i,mb_per_s=31.36,n=2000000i,ns_per_op=733,revision="v1.0.2" 1475695157000000000
go,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=ParsePointWithPrecisionN,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=208i,allocs_per_op=4i,mb_per_s=36.68,n=2000000i,ns_per_op=627,revision="v1.0.2" 1475695157000000000
go,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=ParsePointWithPrecisionU,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=208i,allocs_per_op=4i,mb_per_s=36.15,n=2000000i,ns_per_op=636,revision="v1.0.2" 1475695157000000000
go,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=ParsePointsTagsSorted2,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=240i,allocs_per_op=4i,mb_per_s=53.85,n=2000000i,ns_per_op=947,revision="v1.0.2" 1475695157000000000
go,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=ParsePointsTagsSorted5,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=272i,allocs_per_op=4i,mb_per_s=69.75,n=1000000i,ns_per_op=1189,revision="v1.0.2" 1475695157000000000
go,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=ParsePointsTagsSorted10,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=320i,allocs_per_op=4i,mb_per_s=88.05,n=1000000i,ns_per_op=1624,revision="v1.0.2" 1475695157000000000
go,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=ParsePointsTagsUnSorted2,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=272i,allocs_per_op=5i,mb_per_s=43.69,n=1000000i,ns_per_op=1167,revision="v1.0.2" 1475695157000000000
go,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=ParsePointsTagsUnSorted5,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=336i,allocs_per_op=5i,mb_per_s=50.99,n=1000000i,ns_per_op=1627,revision="v1.0.2" 1475695157000000000
go,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=ParsePointsTagsUnSorted10,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=448i,allocs_per_op=5i,mb_per_s=52.32,n=500000i,ns_per_op=2733,revision="v1.0.2" 1475695157000000000
go,goversion=go1.6.2\ linux/amd64,hwid=c4.large,name=ParseKey,pkg=github.com/influxdata/influxdb/models,procs=2 alloced_bytes_per_op=1030i,allocs_per_op=24i,n=1000000i,ns_per_op=2361,revision="v1.0.2" 1475695157000000000
```
