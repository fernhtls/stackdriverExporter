package prometheusOutput

import (
	monitoring "cloud.google.com/go/monitoring/apiv3"
	"errors"
	"fmt"
	"github.com/fernhtls/stackdriverExporter/stackdriverClient"
	"github.com/fernhtls/stackdriverExporter/utils"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/api/iterator"
	"google.golang.org/genproto/googleapis/api/label"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type PrometheusGaugeMetricDetail struct {
	Name           string
	GaugeMetricVec *prometheus.GaugeVec
}

type PrometheusHistoMetricDetail struct {
	Name           string
	HistoMetricVec *prometheus.HistogramVec
}

type PrometheusGaugeMetric struct {
	MetricsAndInterval         utils.MetricsAndIntervalType
	ResourceTypeGaugeMetricVec map[string]PrometheusGaugeMetricDetail
	StackValueType             metricpb.MetricDescriptor_ValueType
}

type PrometheusHistoMetric struct {
	MetricsAndInterval         utils.MetricsAndIntervalType
	ResourceTypeHistoMetricVec map[string]PrometheusHistoMetricDetail
	StackValueType             metricpb.MetricDescriptor_ValueType
}

var (
	prometheusLogger          *log.Logger
	prometheusMetricsGaugeVec []PrometheusGaugeMetric
	prometheusMetricsHistoVec []PrometheusHistoMetric
)

func init() {
	prometheusLogger = log.New(os.Stdout, "prometheus_server: ", log.LstdFlags)
	prometheusMetricsGaugeVec = make([]PrometheusGaugeMetric, 0)
	prometheusMetricsHistoVec = make([]PrometheusHistoMetric, 0)
}

// OutputConfig : struct for the prometheus config output
type OutputConfig struct {
	ProjectID       string
	BaseHandlerPath string
	Port            int
}

// validates the handler path
func validateHandlerPath(path string) (bool, error) {
	m, err := regexp.MatchString("^/[a-z]+$", path)
	return m, err
}

// ValidateConfig : validating the config for the prometheus output
func (p *OutputConfig) ValidateConfig() error {
	if p.ProjectID == "" {
		return errors.New("project should be configured")
	}
	// validates the HandlePath
	if p.BaseHandlerPath == "" {
		return errors.New("handler path can't be empty")
	}
	// validates the path format
	match, err := validateHandlerPath(p.BaseHandlerPath)
	if !match || err != nil {
		return errors.New("handler path is not valid")
	}
	// valid ports to run the handler
	if p.Port <= 0 {
		return errors.New("port cant be nil or 0")
	}
	return nil
}

// function to return the iterator for adding metrics to prometheus
func getMetricValue(client *stackdriverClient.StackDriverClient,
	metricType string, startTime, endTime *timestamp.Timestamp) (*monitoring.TimeSeriesIterator, error) {
	it, err := client.GetTimeSeriesMetric(metricType, startTime, endTime)
	if err != nil {
		return nil, err
	}
	return it, nil
}

// generates the metric name - type + resource type + label name
func generateMetricName(metricType, resourceType string) (string, error) {
	fullMetricName := ""
	mt := strings.Split(metricType, "/")
	pr := strings.Split(mt[0], ".")
	fullMetricName = strings.Join([]string{pr[0], pr[1], mt[1], mt[2], resourceType}, "_")
	if fullMetricName == "" {
		return "", fmt.Errorf("error on generating name -> metricType: %s, mt: %v", metricType, mt)
	}
	return fullMetricName, nil
}

// gets the metric value for numeric data point
func getMetricValueNumeric(valueType metricpb.MetricDescriptor_ValueType, point *monitoringpb.Point) float64 {
	switch valueType {
	case metricpb.MetricDescriptor_DOUBLE:
		return point.Value.GetDoubleValue()
	case metricpb.MetricDescriptor_INT64:
		return float64(point.Value.GetInt64Value())
	case metricpb.MetricDescriptor_MONEY:
		return point.Value.GetDoubleValue()
	case metricpb.MetricDescriptor_BOOL:
		// boolean metrics can return 1 / 0
		if point.Value.GetBoolValue() {
			return 1.0
		}
		return 0.0
	default:
		return 0.0
	}
}

// Function to get gauge metrics and add to prometheus
func getGaugeMetricsBackground(client *stackdriverClient.StackDriverClient) {
	go func() {
		for {
			for _, gaugeMetric := range prometheusMetricsGaugeVec {
				interval, err := strconv.Atoi(gaugeMetric.MetricsAndInterval.Interval)
				if err != nil {
					prometheusLogger.Fatal(err)
				}
				startTime, endTime, err := utils.GetStartAndEndTimeMinuteInterval(int64(interval))
				if err != nil {
					prometheusLogger.Fatal(err)
				}
				prometheusLogger.Printf("collecting metric %s\n", gaugeMetric.MetricsAndInterval.MetricType)
				it, err := getMetricValue(client, gaugeMetric.MetricsAndInterval.MetricType, startTime, endTime)
				if err != nil {
					prometheusLogger.Fatal(err)
				}
				for {
					resp, err := it.Next()
					if err == iterator.Done {
						break
					}
					if err != nil {
						prometheusLogger.Fatal(err)
					}
					// setting value - getting only latest value to set
					var lastValue float64
					var endTime *timestamp.Timestamp
					for _, p := range resp.GetPoints() {
						if endTime == nil {
							endTime = p.Interval.EndTime
							lastValue = getMetricValueNumeric(gaugeMetric.StackValueType, p)
						}
						if p.Interval.EndTime.AsTime().After(endTime.AsTime()) {
							endTime = p.Interval.EndTime
							lastValue = getMetricValueNumeric(gaugeMetric.StackValueType, p)
						}
					}
					gaugeMetric.ResourceTypeGaugeMetricVec[resp.Resource.Type].GaugeMetricVec.WithLabelValues(
						getMapLabelsValues(resp.Resource.Labels)...).Set(lastValue)
				}
			}
			prometheusLogger.Println("**** ****")
			time.Sleep(1 * time.Minute)
		}
	}()
}

func getMapLabelsValues(mapLabels map[string]string) []string {
	// get keys of the map in alphabetical order
	ko := make([]string, 0)
	for k := range mapLabels {
		ko = append(ko, k)
	}
	sort.Strings(ko)
	l := make([]string, 0)
	for _, lb := range ko {
		l = append(l, mapLabels[lb])
	}
	return l
}

func getStackResourceLabelsKeys(resourceLabels []*label.LabelDescriptor) []string {
	l := make([]string, 0)
	for _, rl := range resourceLabels {
		l = append(l, rl.Key)
	}
	sort.Strings(l)
	return l
}

func registerMetrics(client *stackdriverClient.StackDriverClient, metrics []utils.MetricsAndIntervalType) {
	for _, m := range metrics {
		stackDesc, err := client.GetMetricDescriptor(m.MetricType)
		if err != nil {
			prometheusLogger.Fatal(err)
		}
		resourceTypeGaugeMetricVec := make(map[string]PrometheusGaugeMetricDetail)
		resourceTypeHistoMetricVec := make(map[string]PrometheusHistoMetricDetail)
		for _, resourceType := range stackDesc.MonitoredResourceTypes {
			resourceDesc, err := client.GetMonitoredResourceDescriptor(resourceType)
			if err != nil {
				prometheusLogger.Fatal(err)
			}
			name, err := generateMetricName(m.MetricType, resourceType)
			if err != nil {
				prometheusLogger.Fatal(err)
			}
			switch stackDesc.ValueType {
			case metricpb.MetricDescriptor_STRING:
				prometheusLogger.Println("no string metrics in prometheus")
			case metricpb.MetricDescriptor_DISTRIBUTION:
				pm := prometheus.NewHistogramVec(
					prometheus.HistogramOpts{
						Namespace: "stackdriver",
						Name:      name,
						Help:      strings.Join([]string{stackDesc.Description, resourceDesc.Description}, " "),
					}, getStackResourceLabelsKeys(resourceDesc.Labels))
				resourceTypeHistoMetricVec[resourceType] = PrometheusHistoMetricDetail{
					Name:           name,
					HistoMetricVec: pm,
				}
			default: // all other numeric types
				pm := prometheus.NewGaugeVec(
					prometheus.GaugeOpts{
						Namespace: "stackdriver",
						Name:      name,
						Help:      strings.Join([]string{stackDesc.Description, resourceDesc.Description}, " "),
					}, getStackResourceLabelsKeys(resourceDesc.Labels))
				resourceTypeGaugeMetricVec[resourceType] = PrometheusGaugeMetricDetail{
					Name:           name,
					GaugeMetricVec: pm,
				}
			}
		}
		switch stackDesc.ValueType {
		case metricpb.MetricDescriptor_DISTRIBUTION:
			prometheusMetricsHistoVec = append(prometheusMetricsHistoVec, PrometheusHistoMetric{
				MetricsAndInterval:         m,
				ResourceTypeHistoMetricVec: resourceTypeHistoMetricVec,
				StackValueType:             stackDesc.ValueType,
			})
		default:
			prometheusMetricsGaugeVec = append(prometheusMetricsGaugeVec, PrometheusGaugeMetric{
				MetricsAndInterval:         m,
				ResourceTypeGaugeMetricVec: resourceTypeGaugeMetricVec,
				StackValueType:             stackDesc.ValueType,
			})
		}
	}
	// Registering Gauge Metrics
	for _, m := range prometheusMetricsGaugeVec {
		for _, v := range m.ResourceTypeGaugeMetricVec {
			prometheus.MustRegister(v.GaugeMetricVec)
		}
	}
	// Registering Histogram Metrics
	for _, m := range prometheusMetricsHistoVec {
		for _, v := range m.ResourceTypeHistoMetricVec {
			prometheus.MustRegister(v.HistoMetricVec)
		}
	}
}

// StartServerPrometheusMetrics : starts the http server and process to gather metrics from stackdriver
func (p *OutputConfig) StartServerPrometheusMetrics(metrics []utils.MetricsAndIntervalType) {
	client := stackdriverClient.StackDriverClient{
		ProjectID: p.ProjectID,
	}
	if err := client.InitClient(); err != nil {
		prometheusLogger.Fatal(err)
	}
	// Register all prometheus metrics
	registerMetrics(&client, metrics)
	// Adds data to the metrics
	getGaugeMetricsBackground(&client)
	http.Handle(p.BaseHandlerPath, promhttp.Handler())
	if err := http.ListenAndServe(":"+strconv.Itoa(p.Port), nil); err != nil {
		prometheusLogger.Fatal(err)
	}
}
