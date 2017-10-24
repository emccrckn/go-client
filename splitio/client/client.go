package client

import (
	"github.com/splitio/go-client/splitio/engine/evaluator"
	"github.com/splitio/go-client/splitio/tasks"
	"github.com/splitio/go-client/splitio/util/logging"
)

// SplitClient is the entry-point of the split SDK.
type SplitClient struct {
	Apikey       string
	Logger       logging.LoggerInterface
	LoggerConfig logging.LoggerOptions
	Evaluator    *evaluator.Evaluator
	sync         sdkSync
}

type sdkSync struct {
	splitSync      *tasks.AsyncTask
	segmentSync    *tasks.AsyncTask
	impressionSync *tasks.AsyncTask
	gaugeSync      *tasks.AsyncTask
	countersSync   *tasks.AsyncTask
	latenciesSync  *tasks.AsyncTask
}

// Treatment implements the main functionality of split. Retrieve treatments of a specific feature
// for a certain key and set of attributes
func (c *SplitClient) Treatment(key string, feature string, attributes *map[string]interface{}) string {
	return "CONTROL"
}
