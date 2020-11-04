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
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var prometheusLogger *log.Logger

func init() {
	prometheusLogger = log.New(os.Stdout, "prometheus_server: ", log.LstdFlags)
}

// OutputConfig : struct for the prometheus config output
type OutputConfig struct {
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
func getMetricValue(projectID, metricType string, startTime, endTime *timestamp.Timestamp) (*monitoring.TimeSeriesIterator, error) {
	client := stackdriverClient.StackDriverClient{
		ProjectID:  projectID,
		StartTime:  startTime,
		EndTime:    endTime,
		MetricType: metricType,
	}
	it, err := client.GetTimeSeriesMetric()
	if err != nil {
		return nil, err
	}
	return it, nil
}

// generates the metric name - type + resource type + label name
func generateMetricName(metricType, resourceType string) (string, error) {
	fullMetricName := ""
	mt := strings.Split(metricType, "/")
	fullMetricName = mt[1] + "_" + mt[2] + "_" + resourceType
	if fullMetricName == "" {
		return "", fmt.Errorf("error on generating name -> metricType: %s, mt: %v", metricType, mt)
	}
	return fullMetricName, nil
}

// get the metric value
// todo: check cases of more points in a metrics response
// todo: should work fine for single point metrics
// todo: maybe include logic to return a type SET instead of the default Gauge being used
func getMetricPointValue(valueType metricpb.MetricDescriptor_ValueType, points []*monitoringpb.Point) float64 {
	for i := range points {
		p := points[i]
		switch valueType {
		case metricpb.MetricDescriptor_DOUBLE:
			return p.Value.GetDoubleValue()
		case metricpb.MetricDescriptor_INT64:
			return float64(p.Value.GetInt64Value())
		case metricpb.MetricDescriptor_MONEY:
			return p.Value.GetDoubleValue()
		default:
			return 0.0
		}
	}
	// no points to return
	return 0.0
}

// register the new metric for value types int or double
func registerMetricIntOrDouble(name string, valueType metricpb.MetricDescriptor_ValueType, labels map[string]string,
	points []*monitoringpb.Point) {
	if err := prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace:   "stackdriver",
			Name:        name,
			ConstLabels: labels,
		},
		func() float64 { return getMetricPointValue(valueType, points) },
	)); err != nil {
		if err.Error() != "duplicate metrics collector registration attempted" {
			prometheusLogger.Printf("error on registering metric %s: %v", name, err)
		}
	}
}

// Function to get metrics and add to prometheus
func getMetricsBackground(projectID string, metrics []utils.MetricsAndIntervalType) {
	go func() {
		for {
			for _, metric := range metrics {
				startTime, endTime, err := utils.GetStartAndEndTimeMinuteInterval(metric.Interval)
				if err != nil {
					prometheusLogger.Fatal(err)
				}
				prometheusLogger.Printf("collecting metric %s\n", metric)
				it, err := getMetricValue(projectID, metric.MetricType, startTime, endTime)
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
					name, err := generateMetricName(metric.MetricType, resp.Resource.Type)
					if err != nil {
						prometheusLogger.Println(err)
					}
					switch resp.GetValueType() {
					case metricpb.MetricDescriptor_INT64:
						registerMetricIntOrDouble(name,
							resp.GetValueType(),
							resp.GetResource().Labels,
							resp.GetPoints())
					case metricpb.MetricDescriptor_DOUBLE:
						registerMetricIntOrDouble(
							name,
							resp.GetValueType(),
							resp.GetResource().Labels,
							resp.GetPoints())
					case metricpb.MetricDescriptor_MONEY:
						registerMetricIntOrDouble(
							name,
							resp.GetValueType(),
							resp.GetResource().Labels,
							resp.GetPoints())
					}
				}
			}
			// pushing all intervals to be getting only every minute
			prometheusLogger.Println("**** ****")
			time.Sleep(1 * time.Minute)
		}
	}()
}

// StartServerPrometheusMetrics : starts the http server and process to gather metrics from stackdriver
func (p *OutputConfig) StartServerPrometheusMetrics(projectID string, metrics []utils.MetricsAndIntervalType) {
	getMetricsBackground(projectID, metrics)
	http.Handle(p.BaseHandlerPath, promhttp.Handler())
	if err := http.ListenAndServe(":"+strconv.Itoa(p.Port), nil); err != nil {
		prometheusLogger.Fatal(err)
	}
}
