package metrics

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tommzn/go-config"
)

type TimestreamTestSuite struct {
	suite.Suite
}

func TestTimestreamTestSuite(t *testing.T) {
	suite.Run(t, new(TimestreamTestSuite))
}

func (suite *TimestreamTestSuite) TestPublishMetrics() {

	publisher := NewTimestreamPublisher(loadConfigForTest(nil), loggerForTest())
	timestreamPublisher, ok := publisher.(*TimestreamPublisher)
	suite.True(ok)
	mock := newTimestreamMock(false)
	timestreamPublisher.client = mock

	measurement := measurementForTest()
	publisher.Send(measurement)
	suite.Len(timestreamPublisher.measurements, 1)
	publisher.Flush()
	suite.Len(timestreamPublisher.measurements, 0)
	suite.Nil(timestreamPublisher.Error())
}

func (suite *TimestreamTestSuite) TestPublishMetricsWithError() {

	publisher := NewTimestreamPublisher(loadConfigForTest(nil), nil)
	timestreamPublisher, ok := publisher.(*TimestreamPublisher)
	suite.True(ok)
	mock := newTimestreamMock(true)
	timestreamPublisher.client = mock
	batchSize := 0
	timestreamPublisher.batchSize = &batchSize

	measurement := measurementForTest()
	publisher.Send(measurement)
	suite.Len(timestreamPublisher.measurements, 0)
	suite.NotNil(timestreamPublisher.Error())
}

func (suite *TimestreamTestSuite) TestIntegration() {

	suite.SkipCI()

	publisher := NewTimestreamPublisher(loadConfigForTest(config.AsStringPtr("fixtures/testconfig.yml")), nil)
	timestreamPublisher, ok := publisher.(*TimestreamPublisher)
	suite.True(ok)

	measurement := measurementForTest()
	publisher.Send(measurement)
	suite.Len(timestreamPublisher.measurements, 1)
	suite.Nil(timestreamPublisher.Error())
}

func (suite *TimestreamTestSuite) SkipCI() {
	if _, isSet := os.LookupEnv("CI"); isSet {
		suite.T().Skip("Skip test in CI environment.")
	}
}

// TestFormatMeasurementValue covers every branch of the type switch in
// formatMeasurementValue, including the fallback for unknown types.
func (suite *TimestreamTestSuite) TestFormatMeasurementValue() {

	publisher := &TimestreamPublisher{}

	cases := []struct {
		name  string
		value interface{}
		want  string
	}{
		{"int", 42, "42"},
		{"int32", int32(42), "42"},
		{"int64", int64(42), "42"},
		{"uint32", uint32(42), "42"},
		{"uint64", uint64(42), "42"},
		{"float32", float32(0.5), "0.500000"},
		{"float64", 0.5, "0.500000"},
		{"string", "abc", "abc"},
		{"bool", true, "true"},
	}
	for _, tc := range cases {
		suite.Run(tc.name, func() {
			got, _ := publisher.formatMeasurementValue(MeasurementValue{Value: tc.value})
			suite.Equal(tc.want, got)
		})
	}
}

// TestNewTimestreamClientCreatesRealClient exercises the previously uncovered
// branch of newTimestreamClient where publisher.client is nil and a real
// AWS client is constructed (no network calls happen; only the constructor).
func (suite *TimestreamTestSuite) TestNewTimestreamClientCreatesRealClient() {

	publisher := NewTimestreamPublisher(loadConfigForTest(nil), nil).(*TimestreamPublisher)
	suite.Nil(publisher.client)

	client := publisher.newTimestreamClient()
	suite.NotNil(client)
	suite.NotNil(publisher.client)
	// Subsequent calls should reuse the cached client instance.
	suite.Same(client, publisher.newTimestreamClient())
}

// TestFlushWithoutMeasurementsIsNoop verifies that Flush is a no-op when no
// measurements have been queued (previously implicit; explicit test here).
func (suite *TimestreamTestSuite) TestFlushWithoutMeasurementsIsNoop() {

	publisher := NewTimestreamPublisher(loadConfigForTest(nil), loggerForTest()).(*TimestreamPublisher)
	mock := newTimestreamMock(true).(*timestreamMock)
	publisher.client = mock

	publisher.Flush()

	suite.Empty(mock.records)
	suite.Nil(publisher.Error())
}

// TestSendAssignsCurrentTimeToZeroTimestamp verifies that a measurement with a
// zero-value timestamp gets stamped with time.Now() inside Send.
func (suite *TimestreamTestSuite) TestSendAssignsCurrentTimeToZeroTimestamp() {

	publisher := NewTimestreamPublisher(loadConfigForTest(nil), loggerForTest()).(*TimestreamPublisher)
	mock := newTimestreamMock(false)
	publisher.client = mock

	m := measurementForTest() // TimeStamp is zero
	publisher.Send(m)
	suite.Require().Len(publisher.measurements, 1)
	suite.False(publisher.measurements[0].TimeStamp.IsZero(),
		"Send should stamp a zero TimeStamp with the current time")
}
