package redisdb

import (
	"testing"

	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/datastructures/set"
	"github.com/splitio/go-toolkit/logging"
)

func TestRedisSplitStorage(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	splitStorage := NewRedisSplitStorage("localhost", 6379, 1, "", "testPrefix", logger)

	splitStorage.PutMany([]dtos.SplitDTO{
		dtos.SplitDTO{Name: "split1", ChangeNumber: 1},
		dtos.SplitDTO{Name: "split2", ChangeNumber: 2},
		dtos.SplitDTO{Name: "split3", ChangeNumber: 3},
		dtos.SplitDTO{Name: "split4", ChangeNumber: 4},
	}, 123)

	s1 := splitStorage.Get("split1")
	if s1 == nil || s1.Name != "split1" || s1.ChangeNumber != 1 {
		t.Error("Incorrect split fetched/stored")
	}

	sns := splitStorage.SplitNames()
	snsSet := set.NewSet(sns[0], sns[1], sns[2], sns[3])
	if !snsSet.IsEqual(set.NewSet("split1", "split2", "split3", "split4")) {
		t.Error("Incorrect split names fetched")
		t.Error(sns)
	}

	if splitStorage.Till() != 123 {
		t.Error("Incorrect till")
		t.Error(splitStorage.Till())
	}

	splitStorage.PutMany([]dtos.SplitDTO{
		dtos.SplitDTO{
			Name: "split5",
			Conditions: []dtos.ConditionDTO{
				dtos.ConditionDTO{
					MatcherGroup: dtos.MatcherGroupDTO{
						Matchers: []dtos.MatcherDTO{
							dtos.MatcherDTO{
								UserDefinedSegment: &dtos.UserDefinedSegmentMatcherDataDTO{
									SegmentName: "segment1",
								},
							},
						},
					},
				},
			},
		},
		dtos.SplitDTO{
			Name: "split6",
			Conditions: []dtos.ConditionDTO{
				dtos.ConditionDTO{
					MatcherGroup: dtos.MatcherGroupDTO{
						Matchers: []dtos.MatcherDTO{
							dtos.MatcherDTO{
								UserDefinedSegment: &dtos.UserDefinedSegmentMatcherDataDTO{
									SegmentName: "segment2",
								},
							},
							dtos.MatcherDTO{
								UserDefinedSegment: &dtos.UserDefinedSegmentMatcherDataDTO{
									SegmentName: "segment3",
								},
							},
						},
					},
				},
			},
		},
	}, 123)

	segmentNames := splitStorage.SegmentNames()
	hcSegments := set.NewSet("segment1", "segment2", "segment3")
	if !segmentNames.IsEqual(hcSegments) {
		t.Error("Incorrect segments retrieved")
		t.Error(segmentNames)
		t.Error(hcSegments)
	}

	allSplits := splitStorage.GetAll()
	allNames := set.NewSet()
	for _, split := range allSplits {
		allNames.Add(split.Name)
	}
	if !allNames.IsEqual(set.NewSet("split1", "split2", "split3", "split4", "split5", "split6")) {
		t.Error("GetAll returned incorrect splits")
	}

	for _, name := range []string{"split1", "split2", "split3", "split4", "split5", "split6"} {
		splitStorage.Remove(name)
	}

	allSplits = splitStorage.GetAll()
	if len(allSplits) > 0 {
		t.Error("All splits should have been deleted")
		t.Error(allSplits)
	}
}

func TestSegmentStorage(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	segmentStorage := NewRedisSegmentStorage("localhost", 6379, 1, "", "testPrefix", logger)

	segmentStorage.Put("segment1", set.NewSet("item1", "item2", "item3"), 123)
	segmentStorage.Put("segment2", set.NewSet("item4", "item5", "item6"), 124)

	segment1 := segmentStorage.Get("segment1")
	if segment1 == nil || !segment1.IsEqual(set.NewSet("item1", "item2", "item3")) {
		t.Error("Incorrect segment1")
		t.Error(segment1)
	}

	segment2 := segmentStorage.Get("segment2")
	if segment2 == nil || !segment2.IsEqual(set.NewSet("item4", "item5", "item6")) {
		t.Error("Incorrect segment2")
	}

	if segmentStorage.Till("segment1") != 123 || segmentStorage.Till("segment2") != 124 {
		t.Error("Incorrect till stored")
	}

	segmentStorage.Put("segment1", set.NewSet("item7"), 222)
	segment1 = segmentStorage.Get("segment1")
	if !segment1.IsEqual(set.NewSet("item7")) {
		t.Error("Segment 1 not overwritten correctly")
	}

	if segmentStorage.Till("segment1") != 222 {
		t.Error("segment 1 till not updated correctly")
	}

	segmentStorage.Remove("segment1")
	if segmentStorage.Get("segment1") != nil || segmentStorage.Till("segment1") != -1 {
		t.Error("Segment 1 and it's till value should have been removed")
		t.Error(segmentStorage.Get("segment1"))
		t.Error(segmentStorage.Till("segment1"))
	}
}

func TestImpressionStorage(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	impressionStorage := NewRedisImpressionStorage("localhost", 6379, 1, "", "testPrefix", "instance123", "go-test", logger)

	impressionStorage.Put("feature1", &dtos.ImpressionDTO{
		BucketingKey: "abc",
		ChangeNumber: 123,
		KeyName:      "key1",
		Label:        "label1",
		Time:         111,
		Treatment:    "on",
	})
	impressionStorage.Put("feature1", &dtos.ImpressionDTO{
		BucketingKey: "abc",
		ChangeNumber: 123,
		KeyName:      "key2",
		Label:        "label1",
		Time:         111,
		Treatment:    "off",
	})
	impressionStorage.Put("feature2", &dtos.ImpressionDTO{
		BucketingKey: "abc",
		ChangeNumber: 123,
		KeyName:      "key1",
		Label:        "label1",
		Time:         111,
		Treatment:    "off",
	})
	impressionStorage.Put("feature2", &dtos.ImpressionDTO{
		BucketingKey: "abc",
		ChangeNumber: 123,
		KeyName:      "key2",
		Label:        "label1",
		Time:         111,
		Treatment:    "on",
	})

	if impressionStorage.client.client.Exists(
		"testPrefix.SPLITIO/go-test/instance123/impressions.feature1",
		"testPrefix.SPLITIO/go-test/instance123/impressions.feature2",
	).Val() != 2 {
		t.Error("Keys missing or stored in an incorrect format")
	}

	impressions := impressionStorage.PopAll()

	if len(impressions) != 2 {
		t.Error("Incorrect number of features with impressions fetched")
	}

	var feature1, feature2 dtos.ImpressionsDTO
	if impressions[0].TestName == "feature1" && impressions[1].TestName == "feature2" {
		feature1 = impressions[0]
		feature2 = impressions[1]
	} else if impressions[1].TestName == "feature1" && impressions[0].TestName == "feature2" {
		feature1 = impressions[1]
		feature2 = impressions[0]
	} else {
		t.Error("Incorrect impression testnames!")
		return
	}

	if len(feature1.KeyImpressions) != 2 {
		t.Error("Incorrect number of impressions fetched for feature1")
	}

	if len(feature2.KeyImpressions) != 2 {
		t.Error("Incorrect number of impressions fetched for feature2")
	}

	if impressionStorage.client.client.Exists(
		"testPrefix.SPLITIO/go-test/instance123/impressions.feature1",
		"testPrefix.SPLITIO/go-test/instance123/impressions.feature2",
	).Val() != 0 {
		t.Error("Keys should have been deleted")
	}
}

func TestMetricsStorage(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	metricsStorage := NewRedisMetricsStorage("localhost", 6379, 1, "", "testPrefix", "instance123", "go-test", logger)

	// Gauges

	metricsStorage.PutGauge("g1", 3.345)
	metricsStorage.PutGauge("g2", 4.456)

	if metricsStorage.client.client.Exists(
		"testPrefix.SPLITIO/go-test/instance123/gauge.g1",
		"testPrefix.SPLITIO/go-test/instance123/gauge.g2",
	).Val() != 2 {
		t.Error("Keys or stored in an incorrect format")
	}

	gauges := metricsStorage.PopGauges()

	if len(gauges) != 2 {
		t.Error("Incorrect number of gauges fetched")
		t.Error(gauges)
	}

	var g1, g2 dtos.GaugeDTO
	if gauges[0].MetricName == "g1" {
		g1 = gauges[0]
		g2 = gauges[1]
	} else if gauges[0].MetricName == "g2" {
		g1 = gauges[1]
		g2 = gauges[0]
	} else {
		t.Error("Incorrect gauges names")
		return
	}

	if g1.Gauge != 3.345 || g2.Gauge != 4.456 {
		t.Error("Incorrect gauge values retrieved")
	}

	if metricsStorage.client.client.Exists(
		"testPrefix.SPLITIO/go-test/instance123/gauge.g1",
		"testPrefix.SPLITIO/go-test/instance123/gauge.g2",
	).Val() != 0 {
		t.Error("Gauge keys should have been removed after PopAll() function call")
	}

	// Latencies
	metricsStorage.IncLatency("m1", 13)
	metricsStorage.IncLatency("m1", 13)
	metricsStorage.IncLatency("m1", 13)
	metricsStorage.IncLatency("m1", 1)
	metricsStorage.IncLatency("m1", 1)
	metricsStorage.IncLatency("m2", 1)
	metricsStorage.IncLatency("m2", 2)

	if metricsStorage.client.client.Exists(
		"testPrefix.SPLITIO/go-test/instance123/latency.m1.bucket.13",
		"testPrefix.SPLITIO/go-test/instance123/latency.m1.bucket.1",
		"testPrefix.SPLITIO/go-test/instance123/latency.m2.bucket.1",
		"testPrefix.SPLITIO/go-test/instance123/latency.m2.bucket.2",
	).Val() != 4 {
		t.Error("Keys or stored in an incorrect format")
	}

	latencies := metricsStorage.PopLatencies()
	var m1, m2 dtos.LatenciesDTO
	if latencies[0].MetricName == "m1" {
		m1 = latencies[0]
		m2 = latencies[1]
	} else if latencies[0].MetricName == "m2" {
		m1 = latencies[1]
		m2 = latencies[0]
	} else {
		t.Error("Incorrect latency names")
		return
	}

	if m1.Latencies[13] != 3 || m1.Latencies[1] != 2 {
		t.Error("Incorrect latencies for m1")
	}

	if m2.Latencies[1] != 1 || m2.Latencies[2] != 1 {
		t.Error("Incorrect latencies for m2")
	}

	if metricsStorage.client.client.Exists(
		"testPrefix.SPLITIO/go-test/instance123/latency.m1.bucket.13",
		"testPrefix.SPLITIO/go-test/instance123/latency.m1.bucket.1",
		"testPrefix.SPLITIO/go-test/instance123/latency.m2.bucket.1",
		"testPrefix.SPLITIO/go-test/instance123/latency.m2.bucket.2",
	).Val() != 0 {
		t.Error("Latency keys should have been deleted after PopAll()")
	}

	// Counters
	metricsStorage.IncCounter("count1")
	metricsStorage.IncCounter("count1")
	metricsStorage.IncCounter("count1")
	metricsStorage.IncCounter("count2")
	metricsStorage.IncCounter("count2")
	metricsStorage.IncCounter("count2")
	metricsStorage.IncCounter("count2")
	metricsStorage.IncCounter("count2")
	metricsStorage.IncCounter("count2")

	if metricsStorage.client.client.Exists(
		"testPrefix.SPLITIO/go-test/instance123/count.count1",
		"testPrefix.SPLITIO/go-test/instance123/count.count2",
	).Val() != 2 {
		t.Error("Incorrect counter keys stored in redis")
	}

	counters := metricsStorage.PopCounters()

	var c1, c2 dtos.CounterDTO
	if counters[0].MetricName == "count1" {
		c1 = counters[0]
		c2 = counters[1]
	} else if counters[0].MetricName == "count2" {
		c1 = counters[1]
		c2 = counters[0]
	} else {
		t.Error("Incorrect counters fetched")
	}

	if c1.Count != 3 {
		t.Error("Incorrect count for count1")
	}

	if c2.Count != 6 {
		t.Error("Incorrect count for count2")
	}

	if metricsStorage.client.client.Exists(
		"testPrefix.SPLITIO/go-test/instance123/count.count1",
		"testPrefix.SPLITIO/go-test/instance123/count.count2",
	).Val() != 0 {
		t.Error("Counter keys should have been removed after PopAll()")
	}
}