package env

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

type singletonDynamoDB struct {
	DynamoDB *dynamo.DB
}

var (
	instanceDynamoDB *singletonDynamoDB
	onceDynamoDB     sync.Once
)

// GetDynamoDB func definition
func GetDynamoDB() *dynamo.DB {
	return getDynamoDBInstance().DynamoDB
}

func getDynamoDBInstance() *singletonDynamoDB {
	onceDynamoDB.Do(func() {
		instanceDynamoDB = &singletonDynamoDB{}
		retries := 0

		instanceDynamoDB.DynamoDB = dynamo.New(session.New(), &aws.Config{
			Region:     aws.String("ap-northeast-1"),
			MaxRetries: &retries,
			Endpoint:   &Config.DynamoHost,
			HTTPClient: &http.Client{Timeout: time.Second * 4},
		})
		if instanceDynamoDB.DynamoDB == nil {
			core.Logger.Fatal(errors.New("Failed to connect to DynamoDB"))
		}
	})
	return instanceDynamoDB
}
