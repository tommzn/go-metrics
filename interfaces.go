package metrics

import (
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
)

// Publisher is used to import metrics to different timesteam databases.
type Publisher interface {

	// Send will import a single measurement to a timesteam database. Metrics will maybe be send in batches. Use Flush to force import.
	Send(Measurement)

	// Flush will start importing remaining metrics.
	Flush()

	// Error returns errors from latest pulishing.
	Error() error
}

// TimestreamClient is an interface to write records to timesteam database.
type timestreamClient interface {

	// WriteRecords will import metrics to timestream database
	WriteRecords(input *timestreamwrite.WriteRecordsInput) (*timestreamwrite.WriteRecordsOutput, error)
}
