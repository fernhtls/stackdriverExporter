package main

import (
	"fmt"
	"log"
	"time"

	"github.com/fernhtls/stackdriverExporter/stackdriverClient"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/api/iterator"
)

func main() {
	et := ptypes.TimestampNow()
	st, err := ptypes.TimestampProto(time.Now().Add(-5 * time.Minute))
	if err != nil {
		log.Fatal(err)
	}
	client := stackdriverClient.StackDriverClient{
		ProjectID:  "deployments-metrics",
		StartTime:  st,
		EndTime:    et,
		MetricType: "storage.googleapis.com/storage/total_bytes",
	}
	it, err := client.GetTimeSeriesMetric()
	if err != nil {
		log.Fatal(err)
	}
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Println(fmt.Errorf("could not read time series value: %v", err))
		}
		jm := jsonpb.Marshaler{}
		resJSON, err := jm.MarshalToString(resp)
		fmt.Println(resJSON)

	}
}
