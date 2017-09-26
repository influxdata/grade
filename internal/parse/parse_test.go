package parse_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/influxdata/grade/internal/parse"
)

func TestParse16BenchmarksInPackage(t *testing.T) {
	r := strings.NewReader(`?   	github.com/influxdata/influxdb/services/collectd/test_client	[no test files]
PASS
ok  	github.com/influxdata/influxdb/services/continuous_querier	0.015s
PASS
BenchmarkParse-4	  200000	      5324 ns/op	    1680 B/op	      32 allocs/op
ok  	github.com/influxdata/influxdb/services/graphite	1.152s
PASS
BenchmarkLimitListener	 1000000	      4929 ns/op	     475 B/op	       3 allocs/op
ok  	github.com/influxdata/influxdb/services/httpd	5.131s
`)

	bs, err := parse.ParseMultipleBenchmarks(r)
	if err != nil {
		t.Fatalf("exp no error, got %v", err)
	}

	expBs := map[string][]*parse.Benchmark{
		"github.com/influxdata/influxdb/services/continuous_querier": nil,
		"github.com/influxdata/influxdb/services/graphite": []*parse.Benchmark{
			{
				Name:              "Parse",
				Procs:             4,
				N:                 200000,
				NsPerOp:           5324,
				AllocedBytesPerOp: 1680,
				AllocsPerOp:       32,
				Measured:          parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp,
			},
		},
		"github.com/influxdata/influxdb/services/httpd": []*parse.Benchmark{
			{
				Name:              "LimitListener",
				Procs:             1,
				N:                 1000000,
				NsPerOp:           4929,
				AllocedBytesPerOp: 475,
				AllocsPerOp:       3,
				Measured:          parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp,
			},
		},
	}

	if !reflect.DeepEqual(bs, expBs) {
		t.Fatalf("got %q\nexp %q", bs, expBs)
	}
}

func TestParse17SubbenchmarksInPackage(t *testing.T) {
	r := strings.NewReader(`goos: darwin
goarch: amd64
pkg: github.com/example/append
BenchmarkAppendFloat/Decimal-4         	20000000	        64.8 ns/op	       2 B/op	       4 allocs/op
BenchmarkAppendFloat/Float-4           	10000000	       159 ns/op	       8 B/op	       16 allocs/op
PASS
ok  	github.com/example/append	7.966s
`)

	bs, err := parse.ParseMultipleBenchmarks(r)
	if err != nil {
		t.Fatalf("exp no error, got %v", err)
	}

	expBs := map[string][]*parse.Benchmark{
		"github.com/example/append": []*parse.Benchmark{
			{
				Name:              "AppendFloat/Decimal",
				Procs:             4,
				N:                 20000000,
				NsPerOp:           64.8,
				AllocedBytesPerOp: 2,
				AllocsPerOp:       4,
				Measured:          parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp,
			},
			{
				Name:              "AppendFloat/Float",
				Procs:             4,
				N:                 10000000,
				NsPerOp:           159,
				AllocedBytesPerOp: 8,
				AllocsPerOp:       16,
				Measured:          parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp,
			},
		},
	}

	if !reflect.DeepEqual(bs, expBs) {
		t.Fatalf("got %q\nexp %q", bs, expBs)
	}
}
