package metrics

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite/types"
)

type timestreamMock struct {
	records               []types.Record
	shouldReturnWithError bool
}

func newTimestreamMock(shouldReturnWithError bool) timestreamClient {
	return &timestreamMock{
		records:               []types.Record{},
		shouldReturnWithError: shouldReturnWithError,
	}
}

func (mock *timestreamMock) WriteRecords(ctx context.Context, input *timestreamwrite.WriteRecordsInput, optFns ...func(*timestreamwrite.Options)) (*timestreamwrite.WriteRecordsOutput, error) {
	if mock.shouldReturnWithError {
		return nil, errors.New("an error has occurred")
	}
	for _, record := range input.Records {
		mock.records = append(mock.records, record)
	}
	return &timestreamwrite.WriteRecordsOutput{}, nil
}
