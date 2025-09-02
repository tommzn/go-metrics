package metrics

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite"
)

// Publisher is used to import metrics to different Timestream databases.
type Publisher interface {

	// Send will import a single measurement to a Timestream database.
	// Metrics may be sent in batches. Use Flush to force import.
	Send(Measurement)

	// Flush will start importing remaining metrics.
	Flush()

	// Error returns errors from the latest publishing.
	Error() error
}

// timestreamClient is an interface to write records to Timestream database.
type timestreamClient interface {

	// WriteRecords will import metrics to Timestream database.
	WriteRecords(ctx context.Context, input *timestreamwrite.WriteRecordsInput, optFns ...func(*timestreamwrite.Options)) (*timestreamwrite.WriteRecordsOutput, error)
}
