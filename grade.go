package grade

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client"
)

// Config represents the settings to process benchmarks.
type Config struct {
	// Database is the name of the database into which to store the processed benchmark results.
	Database string

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
}

func (cfg Config) validate() error {
	var msg []string

	if cfg.Database == "" {
		msg = append(msg, "Database cannot be empty")
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
		fmt.Printf("%#v\n", cfg)
		return errors.New(strings.Join(msg, "\n"))
	}

	return nil
}

// Run processes the benchmark output in the given io.Reader and
// converts that data to InfluxDB points to be sent through the provided Client.
func Run(r io.Reader, cl *client.Client, cfg Config) error {
	if err := cfg.validate(); err != nil {
		return err
	}

	return nil
}
