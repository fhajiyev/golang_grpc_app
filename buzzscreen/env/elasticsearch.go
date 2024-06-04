package env

import (
	"net/http"
	"sync"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"gopkg.in/olivere/elastic.v5"
)

// SingletonElasticsearch struct definition
type SingletonElasticsearch struct {
	Elasticsearch *elastic.Client
	//ESContext     context.Context
}

// Tweet struct definition
type Tweet struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

var (
	instanceElasticsearch *SingletonElasticsearch
	onceElasticsearch     sync.Once
)

// GetElasticsearch func definition
func GetElasticsearch() *elastic.Client {
	return getESInstance().Elasticsearch
}

// Search func definition
func (es *SingletonElasticsearch) Search() *elastic.SearchService {
	return instanceElasticsearch.Elasticsearch.Search().Index(Config.ElasticSearch.CampaignIndexName).Type("content_campaign").TimeoutInMillis(1000)
}

func getESInstance() *SingletonElasticsearch {
	onceElasticsearch.Do(func() {

		instanceElasticsearch = &SingletonElasticsearch{}

		core.Logger.Infof("getESInstance() - %s", Config.ElasticSearch.Host)

		var err error

		logger := elasticInfoLogger{}

		httpClient := &http.Client{
			Timeout: time.Second * 3,
		}

		optionsFuncs := []elastic.ClientOptionFunc{
			elastic.SetURL(Config.ElasticSearch.Host),
			elastic.SetHealthcheckTimeout(10 * time.Second),
			elastic.SetSniff(false),
			elastic.SetErrorLog(logger),
			elastic.SetHttpClient(httpClient),
		}

		if IsLocal() || IsDev() {
			optionsFuncs = append(optionsFuncs, elastic.SetTraceLog(logger))
		} else {
			optionsFuncs = append(optionsFuncs, elastic.SetGzip(true))
		}

		instanceElasticsearch.Elasticsearch, err = elastic.NewClient(optionsFuncs...)
		if err != nil {
			core.Logger.WithError(err).Fatal("getESInstance()")
		}
	})
	return instanceElasticsearch
}

type elasticInfoLogger struct{}

// Printf func definition
func (logger elasticInfoLogger) Printf(format string, v ...interface{}) {
	core.Logger.Printf(format, v...)
}
