package tasks

import (
	"encoding/json"
	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/service/api"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/storage/mutexmap"
	"github.com/splitio/go-toolkit/logging"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMetricsSyncTask(t *testing.T) {
	countersRequestReceived := false
	gaugesRequestReceived := false
	latenciesRequestReceived := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Error("Method should be POST")
		}

		body, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			t.Error("Error reading body")
			return
		}

		switch r.URL.Path {
		case "/metrics/times":
			latenciesRequestReceived = true
			var latencies []dtos.LatenciesDTO
			err = json.Unmarshal(body, &latencies)
			if err != nil {
				t.Errorf("Error parsing json: %s", err)
				return
			}
			if latencies[0].MetricName != "metric1" {
				t.Error("Incorrect metric name")
			}
		case "/metrics/counters":
			countersRequestReceived = true
			var counters []dtos.CounterDTO
			err = json.Unmarshal(body, &counters)
			if err != nil {
				t.Errorf("Error parsing json: %s", err)
				return
			}
			if counters[0].MetricName != "counter1" || counters[0].Count != 1 {
				t.Error("Incorrect count received")
			}
		case "/metrics/gauge":
			gaugesRequestReceived = true
			var gauge dtos.GaugeDTO
			err = json.Unmarshal(body, &gauge)
			if err != nil {
				t.Errorf("Error parsing json: %s", err)
				return
			}
			if gauge.MetricName != "g1" || gauge.Gauge != 123 {
				t.Error("Incorrect gauge received")
			}
		default:
			t.Errorf("Incorrect url %s", r.URL.Path)
		}

	}))
	defer ts.Close()

	logger := logging.NewLogger(&logging.LoggerOptions{})
	metricsRecorder := api.NewHTTPMetricsRecorder(
		"",
		&conf.SplitSdkConfig{
			Advanced: &conf.AdvancedConfig{
				EventsURL: ts.URL,
				SdkURL:    ts.URL,
			},
		},
		logger,
	)

	metricsStorage := mutexmap.NewMMMetricsStorage()

	countersTask := NewRecordCountersTask(
		metricsStorage,
		metricsRecorder,
		1,
		"go-0.1",
		"192.168.0.123",
		"machine1",
		logger,
	)
	gaugesTask := NewRecordGaugesTask(
		metricsStorage,
		metricsRecorder,
		1,
		"go-0.1",
		"192.168.0.123",
		"machine1",
		logger,
	)
	latenciesTask := NewRecordLatenciesTask(
		metricsStorage,
		metricsRecorder,
		1,
		"go-0.1",
		"192.168.0.123",
		"machine1",
		logger,
	)

	countersTask.Start()
	gaugesTask.Start()
	latenciesTask.Start()

	if !countersTask.IsRunning() || !gaugesTask.IsRunning() || !latenciesTask.IsRunning() {
		t.Error("Metrics recording tasks should be running")
	}

	metricsStorage.PutGauge("g1", 123)
	metricsStorage.IncLatency("metric1", 5)
	metricsStorage.IncCounter("counter1")
	time.Sleep(time.Second * 2)

	if !countersRequestReceived || !gaugesRequestReceived || !latenciesRequestReceived {
		t.Error("Not all expected requests received")
	}

	countersTask.Stop()
	gaugesTask.Stop()
	latenciesTask.Stop()

	time.Sleep(time.Second * 10)

	if countersTask.IsRunning() || gaugesTask.IsRunning() || latenciesTask.IsRunning() {
		t.Error("Task should be stopped")
	}
}
