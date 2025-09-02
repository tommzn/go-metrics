// Package metrics provides a generic interface to import measurements to timestream databases.
// At the moment AWS Timestream is supported, only.
package metrics

import (
	"time"

	"github.com/tommzn/go-log"
	"github.com/tommzn/go-utils"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// TimestreamPublisher can be used to import metrics to AWS Timestream.
type TimestreamPublisher struct {
	logger       log.Logger
	errorStack   *utils.ErrorStack
	awsConfig    aws.Config
	client       timestreamClient
	database     *string
	table        *string
	batchSize    *int
	measurements []Measurement
}

// Measurement is a single metric which should be imported to a Timestream database.
type Measurement struct {
	MetricName string
	Tags       []MeasurementTag
	Values     []MeasurementValue
	TimeStamp  time.Time
}

// MeasurementTag is a key/value pair for additional metric information.
type MeasurementTag struct {
	Name  string
	Value string
}

// MeasurementValue is a single, named value which can be imported to a Timestream database.
type MeasurementValue struct {
	Name  string
	Value interface{}
}
