package stackdriverClient

import (
	"fmt"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/iterator"
	"testing"
	"time"
)

func TestCreateClient(t *testing.T) {
	// Raises an error as timestamp are equal
	client := StackDriverClient{
		ProjectID:  "deployments-metrics",
	}
	assert.NoError(t, client.InitClient())
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
	}
	err = client.InitClient()
	if err != nil {
		t.Error(err)
	}
	it, err := client.GetTimeSeriesMetric("storage.googleapis.com/storage/total_bytes",
		st, et)
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
