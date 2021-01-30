package stackdriverClient

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/genproto/googleapis/api/metric"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/golang/protobuf/ptypes/timestamp"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
)

// StackDriverClient : struct with the client and all the definitions for accessing stackdriver
// ProjectID	- project id for the connection and extraction of metrics
type StackDriverClient struct {
	ProjectID  string
	client     *monitoring.MetricClient
}

func noMetricTypeError() error {
	return errors.New("should have one or more metric to retrieve")
}

func noStartTimeError() error {
	return errors.New("starttime cannot be empty")
}

func noEndTimeError() error {
	return errors.New("endtime cannot be empty")
}

func invalidIntervalError(startTime *timestamp.Timestamp, endTime *timestamp.Timestamp) error {
	return fmt.Errorf("endtime: %v, cannot be bigger, or equal, than starttime: %v",
		endTime.String(), startTime.String())
}

func (st *StackDriverClient) validateClient() error {
	if st.ProjectID == "" {
		return errors.New("projectid cannot be empty")
	}
	return nil
}

func (st *StackDriverClient) createClient() error {
	// validates the client struct
	if err := st.validateClient(); err != nil {
		return err
	}
	// Creates a new stackdriver client
	// Depends on setting GOOGLE_APPLICATION_CREDENTIALS
	client, err := monitoring.NewMetricClient(context.Background())
	if err != nil {
		return err
	}
	st.client = client
	return nil
}

// InitClient : Initiates a new stackdriver client - only sets client for the project
func (st *StackDriverClient) InitClient() error {
	if err := st.createClient(); err != nil {
		return err
	}
	return nil
}

// GetMetricDescriptor : Gets the descriptor of the metric
func (st *StackDriverClient) GetMetricDescriptor(metricType string) (*metric.MetricDescriptor, error) {
	if metricType == "" {
		return nil, noMetricTypeError()
	}
	descriptor , err := st.client.GetMetricDescriptor(
		context.Background(),
		&monitoringpb.GetMetricDescriptorRequest{
			Name:   "projects/" + st.ProjectID + "/metricDescriptors/" + metricType,
		})
	if err != nil {
		return nil, err
	}
	return descriptor, nil
}

// GetMonitoredResourceDescriptor : Gets the Resource descriptor of the metric
func (st *StackDriverClient) GetMonitoredResourceDescriptor(resourceType string) (*monitoredrespb.MonitoredResourceDescriptor, error) {
	if resourceType == "" {
		return nil, noMetricTypeError()
	}
	resourceDescriptor , err := st.client.GetMonitoredResourceDescriptor(
		context.Background(),
		&monitoringpb.GetMonitoredResourceDescriptorRequest{
			Name:   "projects/" + st.ProjectID + "/monitoredResourceDescriptors/" + resourceType,
		})
	if err != nil {
		return nil, err
	}
	return resourceDescriptor, nil
}

// GetTimeSeriesMetric : Gets the timeseries metrics from stackdriver
func (st *StackDriverClient) GetTimeSeriesMetric(metricType string,
	startTime *timestamp.Timestamp, endTime *timestamp.Timestamp)(*monitoring.TimeSeriesIterator, error) {
	if metricType == "" {
		return nil, noMetricTypeError()
	}
	if startTime == nil {
		return nil, noStartTimeError()
	}
	if endTime == nil {
		return nil, noEndTimeError()
	}
	if endTime.AsTime().Before(startTime.AsTime()) || endTime.AsTime().Equal(startTime.AsTime()) {
		return nil, invalidIntervalError(startTime, endTime)
	}
	it := st.client.ListTimeSeries(context.Background(),
		&monitoringpb.ListTimeSeriesRequest{
			Name:   "projects/" + st.ProjectID,
			Filter: "metric.type = \"" + metricType + "\"",
			Interval: &monitoringpb.TimeInterval{
				StartTime: startTime,
				EndTime:   endTime,
			},
			View: monitoringpb.ListTimeSeriesRequest_FULL,
		})
	return it, nil
}
