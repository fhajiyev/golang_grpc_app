package env

import (
	"strings"
	"sync"

	"github.com/Buzzvil/buzzlib-go/core"
)

var (
	// Config var definition
	Config     = ServerConfigStruct{}
	onceConfig sync.Once
)

type (
	// ServerConfigStruct type definition
	ServerConfigStruct struct {
		Lang string

		Database      *DatabaseConfig
		Redis         *RedisConfig
		StatRedis     *RedisConfig
		ElasticSearch *ElasticsearchConfig

		BuzzconInternalURL string
		SlidejoyURL        string
		InsightURL         string

		AccountsvcURL      string
		AuthsvcURL         string

		ProxyURL            string
		DynamoTableProfile  string
		DynamoTableActivity string
		DynamoTablePoint    string
		DynamoHost          string

		Loggers map[string]*Logger
	}

	// ElasticsearchConfig type definition
	ElasticsearchConfig struct {
		Host              string
		CampaignIndexName string
	}

	// DatabaseConfig type definition
	DatabaseConfig struct {
		Name     string
		User     string
		Password string
		Host     string
		Port     string
		LogMode  bool
	}

	// RedisConfig type definition
	RedisConfig struct {
		Endpoint string
		DB       int
	}

	// Logger type definition
	Logger struct {
		File      string
		Formatter string
	}
)

// IsLocal func definition
func IsLocal() bool {
	return strings.Contains(core.Config.GetString("SERVER_ENV"), "local")
}

// IsTest func definition
func IsTest() bool {
	return strings.Contains(core.Config.GetString("SERVER_ENV"), "test")
}

// IsDev func definition
func IsDev() bool {
	return core.Config.GetString("SERVER_ENV") == "dev"
}

// LoadServerConfig func definition
func LoadServerConfig() {
	onceConfig.Do(func() {
		core.Config.Unmarshal(&Config)
	})
}

//func testKeyValidation(t *testing.T, configMap map[string]interface{}, refType reflect.Type) {
//	if refType.Kind() == reflect.Ptr {
//		refType = refType.Elem()
//	}
//	for key, value := range configMap {
//		fieldValue, found := refType.FieldByName(key)
//		if found == false {
//			t.Fatalf("testKeyValidation() FAIL - Key: %v", key)
//		}
//		if reflect.TypeOf(value) == reflect.TypeOf(configMap) {
//			//t.Logf("testKeyValidation() - value: %v", value)
//			testKeyValidation(t, value.(map[string]interface{}), fieldValue.Type)
//		}
//	}
//}
