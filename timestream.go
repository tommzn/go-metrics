package metrics

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	config "github.com/tommzn/go-config"
	log "github.com/tommzn/go-log"
	utils "github.com/tommzn/go-utils"
)

// NewTimestreamPublisher returns a new metrics publisher for AWS Timestream.
func NewTimestreamPublisher(conf config.Config, logger log.Logger) Publisher {

	if logger == nil {
		logger = log.NewLogger(log.Error, nil, nil)
	}
	batchSize := conf.GetAsInt("aws.timestream.batch_size", nil)
	database := conf.Get("aws.timestream.database", nil)
	table := conf.Get("aws.timestream.table", nil)
	awsConfig := &aws.Config{
		Region: conf.Get("aws.timestream.awsregion", config.AsStringPtr("eu-central-1")),
	}
	return &TimestreamPublisher{
		logger:       logger,
		errorStack:   utils.NewErrorStack(),
		awsConfig:    awsConfig,
		database:     database,
		table:        table,
		batchSize:    batchSize,
		measurements: []Measurement{},
	}
}

// Send will add passed measurement to local queue and trifer Flush if batch size is reached.
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

// SendMeasurements delivers measurements to AWS Timestream. Occurring errors will be colleted and can be accessed by Error method.
func (publisher *TimestreamPublisher) sendMeasurements() {

	if len(publisher.measurements) > 0 {

		publisher.errorStack = utils.NewErrorStack()
		publisher.logger.Infof("Publish %d measurements", len(publisher.measurements))
		records := []*timestreamwrite.Record{}
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
		if _, err := tsClient.WriteRecords(writeRecordsInput); err != nil {
			publisher.logger.Errorf("Timestream write error: %s", err)
			publisher.errorStack.Append(err)
		}
		publisher.measurements = []Measurement{}
	}
}

// ToTimeStreamRecord converts passed measurement to AWS SDK Timestream record.
func (publisher *TimestreamPublisher) toTimeStreamRecord(measurement Measurement) []*timestreamwrite.Record {

	records := []*timestreamwrite.Record{}
	dimensions := publisher.toTimeStreamDimensions(measurement.Tags)
	for _, measurementValue := range measurement.Values {
		measureValue, measureValueType := publisher.formatMeasurementValue(measurementValue)
		records = append(records, &timestreamwrite.Record{
			Dimensions:       dimensions,
			MeasureName:      aws.String(fmt.Sprintf("%s.%s", measurement.MetricName, measurementValue.Name)),
			MeasureValue:     aws.String(measureValue),
			MeasureValueType: aws.String(measureValueType),
			Time:             aws.String(strconv.FormatInt(measurement.TimeStamp.Unix(), 10)),
			TimeUnit:         aws.String("SECONDS"),
		})
	}
	return records
}

// ToTimeStreamDimensions converts passed measurement tag to a timestream dimension.
func (publisher *TimestreamPublisher) toTimeStreamDimensions(tags []MeasurementTag) []*timestreamwrite.Dimension {

	dimensions := []*timestreamwrite.Dimension{}
	for _, tag := range tags {
		dimensions = append(dimensions,
			&timestreamwrite.Dimension{
				Name:  aws.String(tag.Name),
				Value: aws.String(tag.Value),
			},
		)
	}
	return dimensions
}

// FormatMeasurementValue will format passed value depnding on it's type and return a corresponding timestream measurement type.
func (publisher *TimestreamPublisher) formatMeasurementValue(value MeasurementValue) (string, string) {
	switch v := value.Value.(type) {
	case int, uint64, int64, uint32, int32:
		return fmt.Sprintf("%d", v), timestreamwrite.MeasureValueTypeDouble
	case float32, float64:
		return fmt.Sprintf("%f", v), timestreamwrite.MeasureValueTypeDouble
	default:
		return fmt.Sprintf("%s", v), timestreamwrite.MeasureValueTypeVarchar
	}
}

// NewTimestreamClient returns local timestrean client and creates a new one if necessary.
func (publisher *TimestreamPublisher) newTimestreamClient() timestreamClient {
	if publisher.client == nil {
		publisher.client = timestreamwrite.New(session.Must(session.NewSession(publisher.awsConfig)))
	}
	return publisher.client
}
