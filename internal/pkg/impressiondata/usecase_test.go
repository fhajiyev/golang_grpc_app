package impressiondata_test

import (
	"testing"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/impressiondata"
	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/suite"
)

func (ts *UseCaseTestSuite) Test_ImpressionData() {
	var err error
	var impressionData impressiondata.ImpressionData
	err = faker.FakeData(&impressionData)
	ts.NoError(err)

	impressionDataString := ts.useCase.BuildImpressionDataString(impressionData)
	resultImpressionData, err := ts.useCase.ParseImpressionData(impressionDataString)
	ts.NoError(err)
	ts.Equal(impressionData.IFA, resultImpressionData.IFA)
	ts.Equal(impressionData.CampaignID, resultImpressionData.CampaignID)
	ts.Equal(impressionData.DeviceID, resultImpressionData.DeviceID)
	ts.Equal(impressionData.UnitDeviceToken, resultImpressionData.UnitDeviceToken)
	ts.Equal(impressionData.Country, resultImpressionData.Country)
	if impressionData.Gender != nil {
		ts.Equal(*impressionData.Gender, *resultImpressionData.Gender)
	} else {
		ts.Nil(resultImpressionData.Gender)
	}
	if impressionData.YearOfBirth != nil {
		ts.Equal(*impressionData.YearOfBirth, *resultImpressionData.YearOfBirth)
	} else {
		ts.Nil(resultImpressionData.YearOfBirth)
	}
}

var (
	_ suite.SetupTestSuite = &UseCaseTestSuite{}
)

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseTestSuite))
}

type UseCaseTestSuite struct {
	suite.Suite
	useCase impressiondata.UseCase
}

func (ts *UseCaseTestSuite) SetupTest() {
	ts.useCase = impressiondata.NewUseCase()
}
