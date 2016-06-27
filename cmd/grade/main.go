package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/influxdata/grade"
	"github.com/influxdata/influxdb/client"
)

var (
	influxURL string
	unixTime  int64
	insecure  bool

	cfg grade.Config
)

func init() {
	flag.StringVar(&influxURL, "influxurl", "http://localhost:8086", "URL of InfluxDB instance to store benchmark results")
	flag.BoolVar(&insecure, "insecure", false, "Skip SSL verification if set")

	flag.StringVar(&cfg.Database, "database", "benchmarks", "Name of database to store benchmark results")
	flag.StringVar(&cfg.GoVersion, "goversion", "", "Go version used to run benchmarks")
	flag.Int64Var(&unixTime, "timestamp", 0, "Unix epoch timestamp to apply when storing all benchmark results")
	flag.StringVar(&cfg.Revision, "revision", "", "Revision of the repository used to generate benchmark results")
	flag.StringVar(&cfg.HardwareID, "hardwareid", "", "User-specified string to represent the hardware on which the benchmarks were run")
}

func main() {
	flag.Parse()

	cl, err := buildClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating InfluxDB client: %v\n", err)
		os.Exit(1)
		return
	}

	cfg.Timestamp = time.Unix(unixTime, 0)
	if err := grade.Run(os.Stdin, cl, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing benchmarks: %v\n", err)
		os.Exit(1)
		return
	}
}

func buildClient() (*client.Client, error) {
	u, err := url.Parse(influxURL)
	if err != nil {
		return nil, err
	}

	c := client.Config{
		URL:       *u,
		UserAgent: "influxdata.Grade",
		Precision: "s",
		UnsafeSsl: insecure,
	}

	if u.User != nil {
		c.Username = u.User.Username()
		c.Password, _ = u.User.Password()
	}

	return client.NewClient(c)
}
