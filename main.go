package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fernhtls/stackdriverExporter/jsonoutput"
	"github.com/fernhtls/stackdriverExporter/utils"
	cron "github.com/robfig/cron/v3"
)

type metricsListType []string

var projectID string
var parallelism int
var metricsList metricsListType
var outputType string
var outputPath string
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
	flag.StringVar(&outputType, "output_type", "json", "output type for pushing the metrics extracted")
	flag.StringVar(&outputPath, "output_path", "", "optional for when extracting the data to json")
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

func buildJobsOutPut() {
	metricsAndIntervals, err := utils.SetMetricsAndIntervalList(metricsList)
	if err != nil {
		log.Fatal("error on setting metrics list:", err)
	}
	switch outputType {
	case "json":
		if outputPath == "" {
			log.Fatal("should pass a output_path")
		}
		j := jsonoutput.JSONOutput{
			OutputPath: "/tmp",
		}
		err := j.ValidateOutputMethod()
		if err != nil {
			log.Fatal(err)
		}
		err = utils.AddJobs(cronServer, cronLogger, metricsAndIntervals, projectID, &j)
		if err != nil {
			log.Fatal("error on adding jobs to cron server:", err)
		}
	default: // stops process - unrecognized output
		log.Fatal("output type not allowed")
	}
}

func main() {
	flag.Parse()
	validateFlags()
	buildJobsOutPut()
	fmt.Println("")
	fmt.Println("  Project: ", "\t\t", projectID)
	fmt.Println("  Parallelism: ", "\t", parallelism)
	fmt.Println("  Metrics list: ", "\t", metricsList.String())
	fmt.Println("")
	if len(cronServer.Entries()) == 0 {
		log.Fatal("no jobs were added to the cronserver")
	}
	cronServer.Run()
}
