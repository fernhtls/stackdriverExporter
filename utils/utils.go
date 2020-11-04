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

//OutputMethod : interface for the several output methods
type OutputMethod interface {
	GetTimeSeriesMetric(string, string, string)
}

const (
	PrometheusOutput = iota
	JSONOutput
)

// MetricsAndIntervalType : struct with the metric type + the interval
type MetricsAndIntervalType struct {
	MetricType string
	Interval   string
}

// building metric and interval list
func getMetricAndInterval(metricAndIntervalSlice []string, outputType int) (string, string, error) {
	var metricType string
	var intervalMetric string
	if len(metricAndIntervalSlice) > 2 {
		return metricType, intervalMetric, errors.New("more than two arguments passed to generate the metric and interval")
	}
	if len(metricAndIntervalSlice) == 1 {
		switch outputType {
		case PrometheusOutput:
			intervalMetric = "600"
		default:
			intervalMetric = "*/10 * * * *"
		}
	} else {
		switch outputType {
		case PrometheusOutput:
			intervalMetric = metricAndIntervalSlice[1]
		default:
			// checking the cron expression if its valid
			_, err := cronexpr.Parse(metricAndIntervalSlice[1])
			if err != nil {
				return metricType, intervalMetric, err
			}
			intervalMetric = metricAndIntervalSlice[1]
		}
	}
	metricType = metricAndIntervalSlice[0]
	return metricType, intervalMetric, nil
}

// check if metrics are not already in the slice
func checkIfNotInMetricsList(metricType string, metricTypeList []MetricsAndIntervalType) bool {
	found := false
	for _, m := range metricTypeList {
		if m.MetricType == metricType {
			found = true
			break
		}
	}
	return found
}

// SetMetricsAndIntervalList : settting metrics and interval list
func SetMetricsAndIntervalList(metrics []string, outputType int) ([]MetricsAndIntervalType, error) {
	metricsAndInterval := make([]MetricsAndIntervalType, 0)
	for _, metric := range metrics {
		metricType, intervalMetric, err := getMetricAndInterval(strings.Split(metric, "|"), outputType)
		if err != nil {
			return nil, err
		}
		if !checkIfNotInMetricsList(metricType, metricsAndInterval) {
			metricsAndInterval = append(metricsAndInterval, MetricsAndIntervalType{
				MetricType: metricType,
				Interval:   intervalMetric,
			})
		}
	}
	return metricsAndInterval, nil
}

// AddJobs : adds jobs to the cron server
func AddJobs(cronServer *cron.Cron, metricList []MetricsAndIntervalType, projectID string, output OutputMethod) error {
	for _, metricType := range metricList {
		// not passing directly as it passes only the last value to the function calls
		metricTypeMetricType := metricType.MetricType
		metricTypeInterval := metricType.Interval
		_, err := cronServer.AddFunc(metricTypeInterval, func() { output.GetTimeSeriesMetric(projectID, metricTypeMetricType, metricTypeInterval) })
		if err != nil {
			return err
		}
	}
	return nil
}

// GetStartAndEndTimeCronJobs : Returns the start and end time for running a time series
// gets the start time from the interval for the cronjob
func GetStartAndEndTimeCronJobs(cronInterval string) (*timestamppb.Timestamp, *timestamppb.Timestamp, error) {
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

// GetStartAndEndTimeMinuteInterval : Returns the start / end time for an interval from the crontab expression
func GetStartAndEndTimeMinuteInterval(interval int64) (*timestamppb.Timestamp, *timestamppb.Timestamp, error) {
	timeStartFunc := time.Now().Truncate(time.Second)
	endTime, err := ptypes.TimestampProto(timeStartFunc)
	if err != nil {
		return nil, nil, err
	}
	startTime, err := ptypes.TimestampProto(timeStartFunc.Add(-time.Duration(interval) * time.Minute))
	if err != nil {
		return nil, nil, err
	}
	return startTime, endTime, nil
}
