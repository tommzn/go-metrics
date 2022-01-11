package metrics

import (
	"errors"

	"github.com/aws/aws-sdk-go/service/timestreamwrite"
)

type timestreamMock struct {
	records               []*timestreamwrite.Record
	shouldReturnWithError bool
}

func newTimestreamMock(shouldReturnWithError bool) timestreamClient {
	return &timestreamMock{
		records:               []*timestreamwrite.Record{},
		shouldReturnWithError: shouldReturnWithError,
	}
}

func (mock *timestreamMock) WriteRecords(input *timestreamwrite.WriteRecordsInput) (*timestreamwrite.WriteRecordsOutput, error) {
	if mock.shouldReturnWithError {
		return nil, errors.New("An errors has occured!")
	}
	for _, record := range input.Records {
		mock.records = append(mock.records, record)
	}
	return &timestreamwrite.WriteRecordsOutput{}, nil
}
