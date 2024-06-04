package env

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	elastic "gopkg.in/olivere/elastic.v5"
)

// Metrics consists of prometheus metrics.
type Metrics struct {
	*numContentCollector
	// AllocationRequests is counter metric representing total number of allocation requests processed.
	AllocationRequests *prometheus.CounterVec
	// AllocatedContent is histogram metric representing number of content served per allocation request(sampled).
	AllocatedContent *prometheus.HistogramVec
}

type numContentCollector struct {
	gaugeDesc *prometheus.Desc
}

// Describe func definition
func (c *numContentCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.gaugeDesc
}

// Collect func definition
func (c *numContentCollector) Collect(ch chan<- prometheus.Metric) {
	stats, _ := GetContentStats()

	for country, count := range stats {
		ch <- prometheus.MustNewConstMetric(c.gaugeDesc, prometheus.GaugeValue, float64(count), country)
	}
}

func newNumContentCollector() *numContentCollector {
	return &numContentCollector{
		gaugeDesc: prometheus.NewDesc(
			prometheus.BuildFQName("bs", "", "num_content"),
			"Number of content currently indexed to ElasticSearch",
			[]string{"country"}, nil,
		),
	}
}

// NewMetrics initializes prometheus metrics.
func NewMetrics() *Metrics {
	m := &Metrics{
		numContentCollector: newNumContentCollector(),
		AllocationRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "bs",
				Name:      "allocation_requests",
				Help:      "Number of allocation requests.",
			},
			[]string{"country"},
		),
		AllocatedContent: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "bs",
				Name:      "allocated_content",
				Help:      "Number of allocated content per request.",
				Buckets:   []float64{0, 10, 20, 30, 40},
			},
			[]string{"country"},
		),
	}
	prometheus.MustRegister(m.numContentCollector)
	prometheus.MustRegister(m.AllocationRequests)
	prometheus.MustRegister(m.AllocatedContent)
	return m
}

// GetContentStats func definition
func GetContentStats() (map[string]int, error) {
	stats := map[string]int{}
	countryAgg := elastic.NewTermsAggregation().Field("country").Size(100)
	ss := GetElasticsearch().Search().Index(Config.ElasticSearch.CampaignIndexName).Type("content_campaign").Aggregation("country", countryAgg).Size(0).TimeoutInMillis(1000).Preference("_local")

	if IsLocal() {
		ss.Pretty(true)
	}
	sr, err := ss.Do(context.Background())
	if err != nil {
		return stats, err
	}

	aggData, ok := sr.Aggregations.Terms("country")
	if ok {
		for _, bucket := range aggData.Buckets {
			stats[bucket.Key.(string)] = int(bucket.DocCount)
		}
	}
	return stats, nil
}
