package jsonoutput

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/fernhtls/stackdriverExporter/stackdriverClient"
	"github.com/fernhtls/stackdriverExporter/utils"
	"github.com/gogo/protobuf/jsonpb"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func buildFileName(outputPath, metricType string, startTime *timestamppb.Timestamp, endTime *timestamppb.Timestamp) string {

	exp := regexp.MustCompile(`[^\w]`)
	fileName := strings.Join([]string{
		exp.ReplaceAllString(metricType, "-"),
		strconv.FormatInt(startTime.AsTime().Unix(), 10),
		strconv.FormatInt(endTime.AsTime().Unix(), 10)}, "_")
	return filepath.Join(outputPath, fileName+".json")
}

// GetTimeSeriesMetric : writes the metrics capture for the interval in file
func GetTimeSeriesMetric(cronLogger *log.Logger, projectID, metric, cronInterval, outputPath string) {
	startTime, endTime, err := utils.GetStartAndEndTimeJobs(cronInterval)
	if err != nil {
		cronLogger.Println(fmt.Errorf("error on getting start and end time : %v", err))
	}
	cronLogger.Println("getting metrics for type metric", metric, "start:", startTime.AsTime(), "end:", endTime.AsTime())
	client := stackdriverClient.StackDriverClient{
		ProjectID:  projectID,
		StartTime:  startTime,
		EndTime:    endTime,
		MetricType: metric,
	}
	it, err := client.GetTimeSeriesMetric()
	if err != nil {
		cronLogger.Println(fmt.Errorf("error on creating client: %v", err))
	}

	fileName := buildFileName(outputPath, metric, startTime, endTime)
	f, err := os.Create(fileName)
	if err != nil {
		cronLogger.Println(fmt.Errorf("error on creating file to write: %v", err))
	}
	defer f.Close()
	f.Sync()

	cronLogger.Println(fmt.Sprintf("Wrtinting to file: %s", fileName))
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			cronLogger.Println(fmt.Errorf("done error: %v", err))
			break
		}
		if err != nil {
			cronLogger.Println(fmt.Errorf("error retrieving timeseries values: %v", err))
		}
		jm := jsonpb.Marshaler{}
		resJSON, err := jm.MarshalToString(resp)
		if err != nil {
			cronLogger.Println(err)
		}
		_, err = f.WriteString(resJSON + "\n")
		if err != nil {
			cronLogger.Println(fmt.Errorf("error on writing to file : %v", err))
			break
		}
	}
	f.Sync()
}
