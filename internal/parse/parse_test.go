package parse_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/influxdata/grade/internal/parse"
)

func TestParseBenchmarksInPackage(t *testing.T) {
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
				NumCPU:            4,
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
				NumCPU:            1,
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
