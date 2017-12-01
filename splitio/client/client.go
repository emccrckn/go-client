package client

import (
	"errors"
	"runtime/debug"
	"time"

	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/engine/evaluator"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-client/splitio/util/metrics"

	"github.com/splitio/go-toolkit/asynctask"
	"github.com/splitio/go-toolkit/logging"
)

// SplitClient is the entry-point of the split SDK.
type SplitClient struct {
	apikey       string
	cfg          *conf.SplitSdkConfig
	logger       logging.LoggerInterface
	loggerConfig logging.LoggerOptions
	evaluator    evaluator.Interface
	sync         *sdkSync
	impressions  storage.ImpressionStorageProducer
	metrics      storage.MetricsStorageProducer
}

type sdkSync struct {
	splitSync      *asynctask.AsyncTask
	segmentSync    *asynctask.AsyncTask
	impressionSync *asynctask.AsyncTask
	gaugeSync      *asynctask.AsyncTask
	countersSync   *asynctask.AsyncTask
	latenciesSync  *asynctask.AsyncTask
}

func parseKeys(key interface{}) (string, *string, error) {
	var bucketingKey *string
	matchingKey, ok := key.(string)
	bucketingKey = nil
	if !ok {
		ckey, ok := key.(*Key)
		if !ok {
			return "", nil, errors.New("Supplied key is neither a string or a Key struct")
		}
		matchingKey = ckey.MatchingKey
		bucketingKey = &ckey.BucketingKey
	}

	return matchingKey, bucketingKey, nil
}

// Treatment implements the main functionality of split. Retrieve treatments of a specific feature
// for a certain key and set of attributes
func (c *SplitClient) Treatment(key interface{}, feature string, attributes map[string]interface{}) string {
	// Set up a guard deferred function to recover if the SDK starts panicking
	defer func() string {
		if r := recover(); r != nil {
			// At this point we'll only trust that the logger isn't panicking trust
			// that the logger isn't panicking
			c.logger.Error(
				"SDK is panicking with the following error", r, "\n",
				string(debug.Stack()), "\n",
				"Returning CONTROL", "\n",
			)
		}
		return "control"
	}()

	matchingKey, bucketingKey, err := parseKeys(key)
	if err != nil {
		c.logger.Error("Error parsing key: ", err.Error())
		return "control"
	}

	var evaluationResult *evaluator.Result
	var impressionBucketingKey = ""
	if bucketingKey != nil {
		evaluationResult = c.evaluator.Evaluate(matchingKey, bucketingKey, feature, attributes)
		impressionBucketingKey = *bucketingKey
	} else {
		evaluationResult = c.evaluator.Evaluate(matchingKey, &matchingKey, feature, attributes)
	}

	// Store impression
	if c.impressions != nil {
		var label string
		if c.cfg.LabelsEnabled {
			label = evaluationResult.Label
		}
		c.impressions.Put(feature, &dtos.ImpressionDTO{
			BucketingKey: impressionBucketingKey,
			ChangeNumber: evaluationResult.SplitChangeNumber,
			KeyName:      matchingKey,
			Label:        label,
			Treatment:    evaluationResult.Treatment,
			Time:         time.Now().Unix() * 1000, // Convert Unix timestamp to java's ms timestamps
		})
	} else {
		c.logger.Warning("No impression storage set in client. Not sending impressions!")
	}

	// Store latency
	bucket := metrics.Bucket(evaluationResult.EvaluationTimeNs)
	c.metrics.IncLatency("sdk.getTreatment", bucket)

	return evaluationResult.Treatment
}

// Treatments evaluates multiple featers for a single user and set of attributes at once
func (c *SplitClient) Treatments(key interface{}, features []string, attributes map[string]interface{}) map[string]string {
	treatments := make(map[string]string)
	for _, feature := range features {
		treatments[feature] = c.Treatment(key, feature, attributes)
	}
	return treatments
}