package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMetricAndInterval(t *testing.T) {
	metricType, interval, err := getMetricAndInterval([]string{"metrictype", "*/1 * * * *"}, JSONOutput)
	assert.Equal(t, "metrictype", metricType)
	assert.Equal(t, "*/1 * * * *", interval)
	assert.NoError(t, err)
}

func TestGetMetricAndIntervalErrorMoreArguments(t *testing.T) {
	_, _, err := getMetricAndInterval([]string{"metrictype", "*/1 * * * *", "dummy"}, JSONOutput)
	assert.Error(t, err)
}

func TestGetMetricAndIntervalCronError(t *testing.T) {
	_, _, err := getMetricAndInterval([]string{"metrictype", "*/99 * * * *"}, JSONOutput)
	assert.Error(t, err)
}

func TestSetMetricsAndIntervalList(t *testing.T) {
	metricsList := make([]string, 0)
	metricsList = append(metricsList, "storage.googleapis.com/storage/total_bytes|*/10 * * * *")
	metricsList = append(metricsList, "storage.googleapis.com/storage/object_count|*/5 * * * *")
	metricsList = append(metricsList, "storage.googleapis.com/storage/object_count|*/5 * * * *") // not including repeting metrics
	metricsType, err := SetMetricsAndIntervalList(metricsList, JSONOutput)
	assert.Greater(t, len(metricsType), 0)
	assert.Equal(t, len(metricsType), 2)
	assert.NoError(t, err)
	//
	assert.Equal(t, metricsType[0].MetricType, "storage.googleapis.com/storage/total_bytes")
	assert.Equal(t, metricsType[1].MetricType, "storage.googleapis.com/storage/object_count")
	assert.Equal(t, metricsType[0].Interval, "*/10 * * * *")
	assert.Equal(t, metricsType[1].Interval, "*/5 * * * *")
}

func TestSetMetricsAndIntervalListErrorMoreElements(t *testing.T) {
	metricsList := make([]string, 0)
	metricsList = append(metricsList, "storage.googleapis.com/storage/total_bytes|*/10 * * * *")
	metricsList = append(metricsList, "storage.googleapis.com/storage/object_count|*/5 * * * *|Dummy")
	_, err := SetMetricsAndIntervalList(metricsList, JSONOutput)
	assert.Error(t, err)
}

func TestSetMetricsAndIntervalListErrorCronExpression(t *testing.T) {
	metricsList := make([]string, 0)
	metricsList = append(metricsList, "storage.googleapis.com/storage/total_bytes|*/5 * * * *")
	metricsList = append(metricsList, "storage.googleapis.com/storage/object_count|*/99 * * * *")
	_, err := SetMetricsAndIntervalList(metricsList, JSONOutput)
	assert.Error(t, err)
}
