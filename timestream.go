package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite/types"
	"github.com/tommzn/go-config"
	"github.com/tommzn/go-log"
	"github.com/tommzn/go-utils"
)

// NewTimestreamPublisher returns a new metrics publisher for AWS Timestream.
func NewTimestreamPublisher(conf config.Config, logger log.Logger) Publisher {
	if logger == nil {
		logger = log.NewLogger(log.Error, nil, nil)
	}

	batchSize := conf.GetAsInt("aws.timestream.batch_size", nil)
	database := conf.Get("aws.timestream.database", nil)
	table := conf.Get("aws.timestream.table", nil)

	awsRegion := conf.Get("aws.region", config.AsStringPtr("eu-central-1"))
	awsCfg, _ := awsconfig.LoadDefaultConfig(
		context.TODO(),
		awsconfig.WithRegion(*awsRegion),
	)

	return &TimestreamPublisher{
		logger:       logger,
		errorStack:   utils.NewErrorStack(),
		awsConfig:    awsCfg,
		database:     database,
		table:        table,
		batchSize:    batchSize,
		measurements: []Measurement{},
	}
}

// Send will add passed measurement to local queue and trigger Flush if batch size is reached.
func (publisher *TimestreamPublisher) Send(measurement Measurement) {
	publisher.logger.Debugf("Receive measurement: %+v", measurement)
	if measurement.TimeStamp.IsZero() {
		measurement.TimeStamp = time.Now()
	}
	publisher.measurements = append(publisher.measurements, measurement)
	if publisher.batchSizeReached() {
		publisher.logger.Debug("Publishing measurements")
		publisher.Flush()
	}
}

// Flush will start to import all available measurements to AWS Timestream.
func (publisher *TimestreamPublisher) Flush() {
	publisher.sendMeasurements()
}

// Error returns an error if something went wrong during last metric publishing.
func (publisher *TimestreamPublisher) Error() error {
	return publisher.errorStack.AsError()
}

func (publisher *TimestreamPublisher) batchSizeReached() bool {
	return publisher.batchSize == nil ||
		*publisher.batchSize == 0 ||
		len(publisher.measurements) >= *publisher.batchSize
}

// sendMeasurements delivers measurements to AWS Timestream. Occurring errors will be collected and can be accessed by Error method.
func (publisher *TimestreamPublisher) sendMeasurements() {
	if len(publisher.measurements) == 0 {
		return
	}

	publisher.errorStack = utils.NewErrorStack()
	publisher.logger.Infof("Publish %d measurements", len(publisher.measurements))

	var records = []types.Record{}
	for _, measurement := range publisher.measurements {
		newRecords := publisher.toTimeStreamRecord(measurement)
		records = append(records, newRecords...)
	}
	writeRecordsInput := &timestreamwrite.WriteRecordsInput{
		DatabaseName: publisher.database,
		TableName:    publisher.table,
		Records:      records,
	}

	b, _ := json.Marshal(records)
	publisher.logger.Debugf("Publish records: %s", string(b))

	tsClient := publisher.newTimestreamClient()
	_, err := tsClient.WriteRecords(context.Background(), writeRecordsInput)
	if err != nil {
		publisher.logger.Errorf("Timestream write error: %s", err)
		publisher.errorStack.Append(err)
	}
	publisher.measurements = []Measurement{}
}

// toTimeStreamRecord converts passed measurement to AWS SDK Timestream record.
func (publisher *TimestreamPublisher) toTimeStreamRecord(measurement Measurement) []types.Record {
	records := []types.Record{}
	dimensions := publisher.toTimeStreamDimensions(measurement.Tags)

	for _, measurementValue := range measurement.Values {
		measureValue, measureValueType := publisher.formatMeasurementValue(measurementValue)
		records = append(records, types.Record{
			Dimensions:       dimensions,
			MeasureName:      aws.String(fmt.Sprintf("%s.%s", measurement.MetricName, measurementValue.Name)),
			MeasureValue:     aws.String(measureValue),
			MeasureValueType: measureValueType,
			Time:             aws.String(strconv.FormatInt(measurement.TimeStamp.Unix(), 10)),
			TimeUnit:         types.TimeUnitSeconds,
		})
	}
	return records
}

// toTimeStreamDimensions converts passed measurement tags to Timestream dimensions.
func (publisher *TimestreamPublisher) toTimeStreamDimensions(tags []MeasurementTag) []types.Dimension {
	dimensions := make([]types.Dimension, 0, len(tags))
	for _, tag := range tags {
		dimensions = append(dimensions, types.Dimension{
			Name:  aws.String(tag.Name),
			Value: aws.String(tag.Value),
		})
	}
	return dimensions
}

// formatMeasurementValue formats the passed value depending on its type and
// returns a corresponding Timestream measurement type. Integer, unsigned and
// floating-point values are mapped to Double; everything else is stringified
// with fmt.Sprint (safe for arbitrary types, including nil).
func (publisher *TimestreamPublisher) formatMeasurementValue(value MeasurementValue) (string, types.MeasureValueType) {
	switch v := value.Value.(type) {
	case int:
		return strconv.FormatInt(int64(v), 10), types.MeasureValueTypeDouble
	case int32:
		return strconv.FormatInt(int64(v), 10), types.MeasureValueTypeDouble
	case int64:
		return strconv.FormatInt(v, 10), types.MeasureValueTypeDouble
	case uint32:
		return strconv.FormatUint(uint64(v), 10), types.MeasureValueTypeDouble
	case uint64:
		return strconv.FormatUint(v, 10), types.MeasureValueTypeDouble
	case float32:
		// Preserve the previous fmt.Sprintf("%f", ...) formatting (6 decimals).
		return strconv.FormatFloat(float64(v), 'f', 6, 32), types.MeasureValueTypeDouble
	case float64:
		// Preserve the previous fmt.Sprintf("%f", ...) formatting (6 decimals).
		return strconv.FormatFloat(v, 'f', 6, 64), types.MeasureValueTypeDouble
	default:
		return fmt.Sprint(v), types.MeasureValueTypeVarchar
	}
}

// newTimestreamClient returns local timestream client and creates a new one if necessary.
func (publisher *TimestreamPublisher) newTimestreamClient() timestreamClient {
	if publisher.client == nil {
		publisher.client = timestreamwrite.NewFromConfig(publisher.awsConfig)
	}
	return publisher.client
}
