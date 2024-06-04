package trackingdata_test

import (
	"testing"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/trackingdata"
	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/suite"
)

func (ts *UseCaseTestSuite) Test_TrackingData() {
	var err error
	var trackingData *trackingdata.TrackingData
	err = faker.FakeData(&trackingData)
	ts.NoError(err)

	trackingDataString := ts.useCase.BuildTrackingDataString(trackingData)
	resultTrackingData, err := ts.useCase.ParseTrackingData(trackingDataString)
	ts.NoError(err)
	ts.Equal(trackingData.ModelArtifact, resultTrackingData.ModelArtifact)
}

var (
	_ suite.SetupTestSuite = &UseCaseTestSuite{}
)

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseTestSuite))
}

type UseCaseTestSuite struct {
	suite.Suite
	useCase trackingdata.UseCase
}

func (ts *UseCaseTestSuite) SetupTest() {
	ts.useCase = trackingdata.NewUseCase()
}
