package metrics

import (
	config "github.com/tommzn/go-config"
	log "github.com/tommzn/go-log"
)

func loadConfigForTest(fileName *string) config.Config {

	configFile := "fixtures/testconfig.yml"
	if fileName != nil {
		configFile = *fileName
	}
	configLoader := config.NewFileConfigSource(&configFile)
	config, _ := configLoader.Load()
	return config
}

func loggerForTest() log.Logger {
	return log.NewLogger(log.Debug, nil, nil)
}

func measurementForTest() Measurement {
	metric := Measurement{
		MetricName: "test-metric",
		Tags:       []MeasurementTag{},
		Values:     []MeasurementValue{},
	}
	metric.Tags = append(metric.Tags, MeasurementTag{
		Name:  "tag",
		Value: "val",
	})
	metric.Tags = append(metric.Tags, MeasurementTag{
		Name:  "tag2",
		Value: "val2",
	})
	metric.Values = append(metric.Values, MeasurementValue{
		Name:  "count",
		Value: 1,
	})
	metric.Values = append(metric.Values, MeasurementValue{
		Name:  "load",
		Value: 0.53,
	})
	metric.Values = append(metric.Values, MeasurementValue{
		Name:  "enabled",
		Value: true,
	})
	return metric
}
