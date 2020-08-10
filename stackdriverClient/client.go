package stackdriverClient

import (
	"context"
	"errors"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/golang/protobuf/ptypes/timestamp"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

// StackDriverClient : struct with the client and all the definitions for accessing stackdriver
// ProjectID	- project id for the connection and extraction of metrics
// Starttime	- start time (timestamp) for extracting metrics
// EndTime		- end time (timestamp) for extracting metrics
// MetricType	- metric type to extract
type StackDriverClient struct {
	ProjectID  string
	StartTime  *timestamp.Timestamp
	EndTime    *timestamp.Timestamp
	MetricType string
	client     *monitoring.MetricClient
}

func (st *StackDriverClient) validateClient() error {
	if st.ProjectID == "" {
		return errors.New("projectid cannot be empty")
	}
	if st.MetricType == "" {
		return errors.New("should have one or more metric to retrieve")
	}
	if st.EndTime.AsTime().Before(st.StartTime.AsTime()) {
		return errors.New("endtime cannot be bigger than starttime")
	}
	return nil
}

func (st *StackDriverClient) createClient() error {
	// validates the client struct
	err := st.validateClient()
	if err != nil {
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

// GetTimeSeriesMetric : Gets the timeseries metrics from stackdriver
func (st *StackDriverClient) GetTimeSeriesMetric() (*monitoring.TimeSeriesIterator, error) {
	err := st.createClient()
	if err != nil {
		return nil, err
	}
	it := st.client.ListTimeSeries(context.Background(),
		&monitoringpb.ListTimeSeriesRequest{
			Name:   "projects/" + st.ProjectID,
			Filter: "metric.type = \"" + st.MetricType + "\"",
			Interval: &monitoringpb.TimeInterval{
				StartTime: st.StartTime,
				EndTime:   st.EndTime,
			},
			View: monitoringpb.ListTimeSeriesRequest_FULL,
		})
	return it, nil
}
