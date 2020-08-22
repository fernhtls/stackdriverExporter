package jsonoutput

import (
	"errors"
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

// JSONOutput : Struct type for json output
type JSONOutput struct {
	Logger     *log.Logger
	OutputPath string
}

// ValidateOutputMethod : validates the output of the method - does not add the method to job if it fails
func (j *JSONOutput) ValidateOutputMethod() error {
	if j.OutputPath == "" {
		return errors.New("OutputPath can't be blank")
	}
	pathExists, err := os.Stat(j.OutputPath)
	if err != nil && pathExists == nil {
		return fmt.Errorf("path %s does not exist: %v", j.OutputPath, err)
	}
	if !pathExists.IsDir() {
		return fmt.Errorf("path %s is not a directory", j.OutputPath)
	}
	return nil
}

func (j *JSONOutput) buildFileName(metricType string, startTime *timestamppb.Timestamp, endTime *timestamppb.Timestamp) string {
	exp := regexp.MustCompile(`[^\w]`)
	fileName := strings.Join([]string{
		exp.ReplaceAllString(metricType, "-"),
		strconv.FormatInt(startTime.AsTime().Unix(), 10),
		strconv.FormatInt(endTime.AsTime().Unix(), 10)}, "_")
	return filepath.Join(j.OutputPath, fileName+".json")
}

// GetTimeSeriesMetric : writes the metrics capture for the interval in file
func (j *JSONOutput) GetTimeSeriesMetric(projectID, metric, cronInterval string) {
	startTime, endTime, err := utils.GetStartAndEndTimeJobs(cronInterval)
	if err != nil {
		j.Logger.Println(fmt.Errorf("error on getting start and end time : %v", err))
	}
	j.Logger.Println("getting metrics for type metric", metric, "start:", startTime.AsTime(), "end:", endTime.AsTime())
	client := stackdriverClient.StackDriverClient{
		ProjectID:  projectID,
		StartTime:  startTime,
		EndTime:    endTime,
		MetricType: metric,
	}
	it, err := client.GetTimeSeriesMetric()
	if err != nil {
		j.Logger.Println(fmt.Errorf("error on creating client: %v", err))
	}

	fileName := j.buildFileName(metric, startTime, endTime)
	f, err := os.Create(fileName)
	if err != nil {
		j.Logger.Println(fmt.Errorf("error on creating file to write: %v", err))
	}
	defer f.Close()
	j.Logger.Println(fmt.Sprintf("Wrtinting to file: %s", fileName))
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			j.Logger.Println(fmt.Errorf("done error: %v", err))
			break
		}
		if err != nil {
			j.Logger.Println(fmt.Errorf("error retrieving timeseries values: %v", err))
		}
		jm := jsonpb.Marshaler{}
		resJSON, err := jm.MarshalToString(resp)
		if err != nil {
			j.Logger.Println(err)
		}
		_, err = f.WriteString(resJSON + "\n")
		if err != nil {
			j.Logger.Println(fmt.Errorf("error on writing to file : %v", err))
			break
		}
	}
	f.Sync()
}
