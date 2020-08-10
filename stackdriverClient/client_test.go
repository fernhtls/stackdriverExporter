package stackdriverClient

import (
	"fmt"
	"testing"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/iterator"
)

func TestCreateClientErrors(t *testing.T) {
	// Raises an error as timestamp are equal
	timeSame := ptypes.TimestampNow()
	clientErrorTimestamp := StackDriverClient{
		ProjectID:  "deployments-metrics",
		StartTime:  timeSame,
		EndTime:    timeSame,
		MetricType: "storage.googleapis.com/storage/total_bytes",
	}
	err := clientErrorTimestamp.createClient()
	assert.Error(t, err)
	// Raises an error as timestamp are equal
	clientErroremptyTimestamp := StackDriverClient{
		ProjectID:  "deployments-metrics",
		StartTime:  timeSame,
		MetricType: "storage.googleapis.com/storage/total_bytes",
	}
	err = clientErroremptyTimestamp.createClient()
	assert.Error(t, err)
}

func TestCreateClient(t *testing.T) {
	et := ptypes.TimestampNow()
	st, err := ptypes.TimestampProto(time.Now().Add(-5 * time.Minute))
	if err != nil {
		t.Error(err)
	}
	client := StackDriverClient{
		ProjectID:  "deployments-metrics",
		StartTime:  st,
		EndTime:    et,
		MetricType: "storage.googleapis.com/storage/total_bytes",
	}
	err = client.createClient()
	assert.NoError(t, err)
}

func TestGetTimeSeriesMetric(t *testing.T) {
	// Depends on setting GOOGLE_APPLICATION_CREDENTIALS
	// or gcloud auth application-default login
	et := ptypes.TimestampNow()
	st, err := ptypes.TimestampProto(time.Now().Add(-15 * time.Minute))
	if err != nil {
		t.Error(err)
	}
	client := StackDriverClient{
		ProjectID:  "deployments-metrics",
		StartTime:  st,
		EndTime:    et,
		MetricType: "storage.googleapis.com/storage/total_bytes",
	}
	err = client.createClient()
	if err != nil {
		t.Error(err)
	}
	it, err := client.GetTimeSeriesMetric()
	if err != nil {
		t.Error(err)
	}
	respsJSON := make([]string, 0)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Println(fmt.Errorf("could not read time series value: %v", err))
		}
		jm := jsonpb.Marshaler{}
		respJSON, err := jm.MarshalToString(resp)
		if err != nil {
			t.Error(err)
		}
		respsJSON = append(respsJSON, respJSON)
	}
	assert.GreaterOrEqual(t, len(respsJSON), 0)
	fmt.Println(respsJSON)
}
