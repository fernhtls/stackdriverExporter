package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	json "github.com/fernhtls/stackdriverExporter/jsonoutput"
	"github.com/fernhtls/stackdriverExporter/utils"
	cron "github.com/robfig/cron/v3"
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

func main() {
	flag.Parse()
	validateFlags()
	metricsAndIntervals, err := utils.SetMetricsAndIntervalList(metricsList)
	if err != nil {
		log.Fatal("error on setting metrics list:", err)
	}
	err = utils.AddJobs(cronServer, cronLogger, metricsAndIntervals, projectID, "/tmp", json.GetTimeSeriesMetric)
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
