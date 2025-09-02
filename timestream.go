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

// toTimeStreamDimensions converts passed measurement tag to a Timestream dimension.
func (publisher *TimestreamPublisher) toTimeStreamDimensions(tags []MeasurementTag) []types.Dimension {
	dimensions := []types.Dimension{}
	for _, tag := range tags {
		dimensions = append(dimensions,
			types.Dimension{
				Name:  aws.String(tag.Name),
				Value: aws.String(tag.Value),
			},
		)
	}
	return dimensions
}

// formatMeasurementValue will format passed value depending on its type and return a corresponding Timestream measurement type.
func (publisher *TimestreamPublisher) formatMeasurementValue(value MeasurementValue) (string, types.MeasureValueType) {
	switch v := value.Value.(type) {
	case int, int32, int64, uint32, uint64:
		return fmt.Sprintf("%d", v), types.MeasureValueTypeDouble
	case float32, float64:
		return fmt.Sprintf("%f", v), types.MeasureValueTypeDouble
	default:
		return fmt.Sprintf("%s", v), types.MeasureValueTypeVarchar
	}
}

// newTimestreamClient returns local timestream client and creates a new one if necessary.
func (publisher *TimestreamPublisher) newTimestreamClient() timestreamClient {
	if publisher.client == nil {
		publisher.client = timestreamwrite.NewFromConfig(publisher.awsConfig)
	}
	return publisher.client
}
