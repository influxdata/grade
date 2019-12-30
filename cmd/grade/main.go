package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/influxdata/grade"
	client "github.com/influxdata/influxdb/client/v2"
)

var (
	influxURL string
	unixTime  int64
	insecure  bool

	cfg grade.Config
)

func init() {
	flag.StringVar(&influxURL, "influxurl", "http://localhost:8086", "URL of InfluxDB instance to store benchmark results (set to empty string to print line protocol to stdout)")
	flag.BoolVar(&insecure, "insecure", false, "Skip SSL verification if set")

	flag.StringVar(&cfg.Database, "database", "benchmarks", "Name of database to store benchmark results")
	flag.StringVar(&cfg.Measurement, "measurement", "go", "Name of measurement to store benchmark results")
	flag.StringVar(&cfg.GoVersion, "goversion", "", "Go version used to run benchmarks")
	flag.Int64Var(&unixTime, "timestamp", 0, "Unix epoch timestamp (in seconds) to apply when storing all benchmark results")
	flag.StringVar(&cfg.Revision, "revision", "", "Revision of the repository used to generate benchmark results")
	flag.StringVar(&cfg.HardwareID, "hardwareid", "", "User-specified string to represent the hardware on which the benchmarks were run")
	flag.StringVar(&cfg.Branch, "branch", "", "Branch of the repository used to generate benchmark results. The flag is optional and can be omitted")
}

func main() {
	flag.Parse()
	cfg.Timestamp = time.Unix(unixTime, 0)

	points, err := grade.Points(os.Stdin, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing benchmarks: %v\n", err)
		os.Exit(1)
	}

	if influxURL == "" {
		// Dry run requested.
		for _, p := range points.Points() {
			fmt.Println(p.String())
		}
	} else {
		cl, err := buildClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating InfluxDB client: %v\n", err)
			os.Exit(1)
		}
		defer cl.Close()

		if err := cl.Write(points); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing benchmark data to InfluxDB: %v\n", err)
			os.Exit(1)
		}
	}
}

func buildClient() (client.Client, error) {
	u, err := url.Parse(influxURL)
	if err != nil {
		return nil, err
	}

	c := client.HTTPConfig{
		Addr:               influxURL,
		UserAgent:          "influxdata.Grade",
		InsecureSkipVerify: insecure,
	}

	if u.User != nil {
		c.Username = u.User.Username()
		c.Password, _ = u.User.Password()
	}

	return client.NewHTTPClient(c)
}
