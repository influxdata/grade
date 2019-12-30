package grade

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/grade/internal/parse"
	client "github.com/influxdata/influxdb/client/v2"
)

// Config represents the settings to process benchmarks.
type Config struct {
	// Database is the name of the database into which to store the processed benchmark results.
	Database string

	// Measurement is the name of the measurement into which to store the processed benchmark results.
	Measurement string

	// GoVersion is the tag value to use to indicate which version of Go was used for the benchmarks that have run.
	GoVersion string

	// Timestamp is the time to use when recording all of the benchmark results,
	// and is typically the timestamp of the commit used for the benchmark.
	Timestamp time.Time

	// Revision is the tag value to use to indicate which revision of the repository was used for the benchmarks that have run.
	// Feel free to use a SHA, tag name, or whatever will be useful to you when querying.
	Revision string

	// HardwareID is a user-specified string to represent the hardware on which the benchmarks have run.
	HardwareID string

	// Branch is the tag value to use to indicate which branch of the repository was used for the benchmarks that have run.
	// The tag is optional and can be omitted.
	Branch string
}

func (cfg Config) validate() error {
	var msg []string

	if cfg.Database == "" {
		msg = append(msg, "Database cannot be empty")
	}

	if cfg.Measurement == "" {
		msg = append(msg, "Measurement cannot be empty")
	}

	if cfg.GoVersion == "" {
		msg = append(msg, "Go version cannot be empty")
	}

	if cfg.Timestamp.Unix() <= 0 {
		msg = append(msg, "Timestamp must be greater than zero")
	}

	if cfg.Revision == "" {
		msg = append(msg, "Revision cannot be empty")
	}

	if cfg.HardwareID == "" {
		msg = append(msg, "Hardware ID cannot be empty")
	}

	if len(msg) > 0 {
		return errors.New(strings.Join(msg, "\n"))
	}

	return nil
}

// Points parses the benchmark output from r and creates a batch of points using cfg.
func Points(r io.Reader, cfg Config) (client.BatchPoints, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	benchset, err := parse.ParseMultipleBenchmarks(r)
	if err != nil {
		return nil, err
	}

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Precision: "s",
		Database:  cfg.Database,
	})
	if err != nil {
		return nil, err
	}

	for pkg, bs := range benchset {
		for _, b := range bs {
			tags := map[string]string{
				"goversion": cfg.GoVersion,
				"hwid":      cfg.HardwareID,
				"pkg":       pkg,
				"procs":     strconv.Itoa(b.Procs),
				"name":      b.Name,
			}
			if cfg.Branch != "" {
				tags["branch"] = cfg.Branch
			}
			p, err := client.NewPoint(
				cfg.Measurement,
				tags,
				makeFields(b, cfg.Revision),
				cfg.Timestamp,
			)
			if err != nil {
				return nil, err
			}

			bp.AddPoint(p)
		}
	}

	return bp, nil
}

func makeFields(b *parse.Benchmark, revision string) map[string]interface{} {
	f := make(map[string]interface{}, 6)

	f["revision"] = revision
	f["n"] = b.N

	if (b.Measured & parse.NsPerOp) != 0 {
		f["ns_per_op"] = b.NsPerOp
	}
	if (b.Measured & parse.MBPerS) != 0 {
		f["mb_per_s"] = b.MBPerS
	}
	if (b.Measured & parse.AllocedBytesPerOp) != 0 {
		f["alloced_bytes_per_op"] = int64(b.AllocedBytesPerOp)
	}
	if (b.Measured & parse.AllocsPerOp) != 0 {
		f["allocs_per_op"] = int64(b.AllocsPerOp)
	}

	return f
}
