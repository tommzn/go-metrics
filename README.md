[![Go Reference](https://pkg.go.dev/badge/github.com/tommzn/go-metrics.svg)](https://pkg.go.dev/github.com/tommzn/go-metrics)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/tommzn/go-metrics)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/tommzn/go-metrics)
[![Go Report Card](https://goreportcard.com/badge/github.com/tommzn/go-metrics)](https://goreportcard.com/report/github.com/tommzn/go-metrics)
[![Actions Status](https://github.com/tommzn/go-metrics/actions/workflows/go.pkg.auto-ci.yml/badge.svg)](https://github.com/tommzn/go-metrics/actions)

# Timestream DB Client
Client to import measurements to timestream databases.
## Supported Databases/Services
- AWS Timestream

# AWS Timestream
Run NewTimestreamPublisher to create a new client for AWS Timestream. You've to pass a config which defines region, database, table name and batch size.
If you don't pass a logger a new stdout logger will be created. See https://github.com/tommzn/go-log for more details about used logger.
```golang
package main

import (
    "fmt"

    metrics "github.com/tommzn/go-metrics"
    config "github.com/tommzn/go-config"
)

func main() {

    // Assumes a config.yml file in current folter.
    conf, err := config.NewConfigSource().Load()
    if err != nil {
        panic(err)
    }
    publisher := metrics.NewTimestreamPublisher(conf. nil)
    measurement := Measurement{
		MetricName: "test-metric",
		Tags: []MeasurementTag{MeasurementTag{
			Name:  "host",
			Value: "ip-10.0.1.12",
		}},
		Values: []MeasurementValue{MeasurementValue{
			Name:  "load",
			Value: 0.53,
		}},
	}

    // If batch size is not set or 0 this will directly import passed measurement to AWS Timestream.
    publisher.Send(measurement)
    // To ensure import of a aavailable measurements you've to flush the publisher.
    publisher.Flush()

    // Any errors that occur when importing measurements are collected.
    if err := publisher.Error(); err != nil {
        fmt.Println(err)
    }
}
```

## Config
Following examples show a config with all available values.
```yaml
aws:
  timestream:
    awsregion: eu-west-1
    database: timestream-db
    table: data
    batch_size: 10
```

## AWS Credentials
Credentials for AWS access have to be setup separately as described in [AWS SDK Docs](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials).
