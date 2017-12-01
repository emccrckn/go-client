package service

import "github.com/splitio/go-client/splitio/service/dtos"

// SplitFetcher interface to be implemented by Split Fetchers
type SplitFetcher interface {
	Fetch(changeNumber int64) (*dtos.SplitChangesDTO, error)
}

// SegmentFetcher interface to be implemented by Split Fetchers
type SegmentFetcher interface {
	Fetch(name string, changeNumber int64) (*dtos.SegmentChangesDTO, error)
}

// ImpressionsRecorder interface to be implemented by Impressions loggers
type ImpressionsRecorder interface {
	Record(impressions []dtos.ImpressionsDTO, sdkVersion string, machineIP string, machineName string) error
}

// MetricsRecorder interface to be implemented by Metrics loggers
type MetricsRecorder interface {
	RecordLatencies(latencies []dtos.LatenciesDTO, sdkVersion string, machineIP string, machineName string) error
	RecordCounters(counters []dtos.CounterDTO, sdkVersion string, machineIP string, machineName string) error
	RecordGauge(gauge dtos.GaugeDTO, sdkVersion string, machineIP string, machineName string) error
}