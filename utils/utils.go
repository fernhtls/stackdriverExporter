package utils

import (
	"errors"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/gorhill/cronexpr"
	cron "github.com/robfig/cron/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MetricsAndIntervalType : struct with the metric type + the interval
type MetricsAndIntervalType struct {
	MetricType string
	Interval   string
}

// building metric and interval list
func getMetricAndInterval(metricAndIntervalSlice []string) (string, string, error) {
	var metricType string
	var intervalMetric string
	if len(metricAndIntervalSlice) > 2 {
		return metricType, intervalMetric, errors.New("more than two arguments passed to generate the metric and interval")
	}
	// checking the cron expression if its valid
	_, err := cronexpr.Parse(metricAndIntervalSlice[1])
	if err != nil {
		return metricType, intervalMetric, err
	}
	if len(metricAndIntervalSlice) == 1 {
		// only contain the metric
		// adding 5 min interval
		intervalMetric = "*/5 * * * *"
	} else {
		intervalMetric = metricAndIntervalSlice[1]
	}
	metricType = metricAndIntervalSlice[0]
	return metricType, intervalMetric, nil
}

// SetMetricsAndIntervalList : settting metrics and interval list
func SetMetricsAndIntervalList(metrics []string) ([]MetricsAndIntervalType, error) {
	metricsAndInterval := make([]MetricsAndIntervalType, 0)
	for _, metric := range metrics {
		metricType, intervalMetric, err := getMetricAndInterval(strings.Split(metric, "|"))
		if err != nil {
			return nil, err
		}
		metricsAndInterval = append(metricsAndInterval, MetricsAndIntervalType{
			MetricType: metricType,
			Interval:   intervalMetric,
		})
	}
	return metricsAndInterval, nil
}

// AddJobs : adds jobs to the cron server
func AddJobs(cronServer *cron.Cron, metricList []MetricsAndIntervalType, getMetric func(string, string)) error {
	for _, metricType := range metricList {
		// not passing directly as it passes only the last value to the function calls
		metricTypeMetricType := metricType.MetricType
		metricTypeInterval := metricType.Interval
		_, err := cronServer.AddFunc(metricTypeInterval, func() { getMetric(metricTypeMetricType, metricTypeInterval) })
		if err != nil {
			return err
		}
	}
	return nil
}

// GetStartAndEndTimeJobs : Returns the start and end time for running a time series
func GetStartAndEndTimeJobs(cronInterval string) (*timestamppb.Timestamp, *timestamppb.Timestamp, error) {
	timeStartFunc := time.Now().Truncate(time.Second)
	e, err := cronexpr.Parse(cronInterval)
	if err != nil {
		return nil, nil, err
	}
	elapsed := e.NextN(timeStartFunc, 1)[0].Sub(timeStartFunc)
	endTime, err := ptypes.TimestampProto(timeStartFunc)
	if err != nil {
		return nil, nil, err
	}
	startTime, err := ptypes.TimestampProto(timeStartFunc.Add(-elapsed.Truncate(time.Minute)))
	if err != nil {
		return nil, nil, err
	}
	return startTime, endTime, nil
}