package metrics

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	config "github.com/tommzn/go-config"
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

	publisher := NewTimestreamPublisher(loadConfigForTest(config.AsStringPtr("fixtures/testconfig01.yml")), nil)
	timestreamPublisher, ok := publisher.(*TimestreamPublisher)
	suite.True(ok)

	measurement := measurementForTest()
	publisher.Send(measurement)
	suite.Len(timestreamPublisher.measurements, 0)
	suite.Nil(timestreamPublisher.Error())
}

func (suite *TimestreamTestSuite) SkipCI() {
	if _, isSet := os.LookupEnv("CI"); isSet {
		suite.T().Skip("Skip test in CI environment.")
	}
}
