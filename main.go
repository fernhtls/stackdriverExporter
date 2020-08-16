package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fernhtls/stackdriverExporter/stackdriverClient"
	"github.com/fernhtls/stackdriverExporter/utils"
	"github.com/golang/protobuf/jsonpb"
	cron "github.com/robfig/cron/v3"
	"google.golang.org/api/iterator"
)

type metricsListType []string

var projectID string
var parallelism int
var metricsList metricsListType
var cronServer *cron.Cron
var cronLogger *log.Logger

func (m *metricsListType) String() string {
	return strings.Join(*m, ", ")
}

func (m *metricsListType) Set(value string) error {
	*m = append(*m, value)
	return nil
}

func init() {
	// Definitions of all flags to run and get from command line
	flag.Usage = func() {
		fmt.Println("Help / Usage of stackdriverExporter :")
		fmt.Println("")
		flag.PrintDefaults()
		fmt.Println("")
	}
	flag.StringVar(&projectID, "project_id", "", "gcp project id to connect and extract the metrics")
	textMetricFlag := "Metric types to extract (pass --metric_type multiple time to extract multiple metrics)"
	textMetricFlag += "\nAfter a pipe character (\"|\"), add as well the interval to collect the metric as a cron expression like \"5/* * * * *\""
	textMetricFlag += "\nExample: --metric_type \"storage.googleapis.com/storage/total_bytes|*/5 * * * *\" "
	flag.Var(&metricsList, "metric_type", textMetricFlag)
	cronLogger = log.New(os.Stdout, "cron_server: ", log.LstdFlags)
	// New cron server
	cronServer = cron.New(
		cron.WithLogger(
			cron.VerbosePrintfLogger(cronLogger)))
}

// validation of flags being passed on command line
func validateFlags() {
	if projectID == "" {
		log.Fatal("project id is mandatory")
	}
	if len(metricsList) == 0 {
		log.Fatal("provide at least one metrics to be extracted")
	}
}

// calls get metrics
// todo: all these methods will be pushed to its own way of pushign the data back
// todo: add interval from job
func getTimeSeriesMetric(metric string, cronInterval string) {
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
		cronLogger.Println(resJSON)
	}
}

func main() {
	flag.Parse()
	validateFlags()
	metricsAndIntervals, err := utils.SetMetricsAndIntervalList(metricsList)
	if err != nil {
		log.Fatal("error on setting metrics list:", err)
	}
	err = utils.AddJobs(cronServer, metricsAndIntervals, getTimeSeriesMetric)
	if err != nil {
		log.Fatal("error on adding jobs to cron server:", err)
	}
	fmt.Println("")
	fmt.Println("  Project: ", "\t\t", projectID)
	fmt.Println("  Parallelism: ", "\t", parallelism)
	fmt.Println("  Metrics list: ", "\t", metricsList.String())
	fmt.Println("")
	cronServer.Run()
}
