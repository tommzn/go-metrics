// Package metrics provides a generic interface to import measurements to timestream databases. At the moment AWS Timestream is supported, only.
package metrics

import (
	"time"

	log "github.com/tommzn/go-log"
	utils "github.com/tommzn/go-utils"

	"github.com/aws/aws-sdk-go/aws"
)

// TimestreamPublisher can be used to import metics to AWS Timestream.
type TimestreamPublisher struct {
	logger       log.Logger
	errorStack   *utils.ErrorStack
	awsConfig    *aws.Config
	client       timestreamClient
	database     *string
	table        *string
	batchSize    *int
	measurements []Measurement
}

// Measurement a single metrics which should be imported to a timestream database.
type Measurement struct {
	MetricName string
	Tags       []MeasurementTag
	Values     []MeasurementValue
	TimeStamp  time.Time
}

// MeasurementTag is a key/value apir for additional metric information.
type MeasurementTag struct {
	Name  string
	Value string
}

// MeasurementValue is a single, named value which can be imported to a timestream database.
type MeasurementValue struct {
	Name  string
	Value interface{}
}
