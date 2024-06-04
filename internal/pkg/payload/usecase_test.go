package payload_test

import (
	"testing"
	"time"

	"github.com/bxcodec/faker"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/payload"

	"github.com/stretchr/testify/suite"
)

func (ts *UseCaseTestSuite) TestPayloadSetUnitID() {
	var p *payload.Payload
	ts.NoError(faker.FakeData(&p))

	unitID := int64(100000045)
	sdkVersion := 3500
	p.SetUnitID(unitID, sdkVersion)
	ts.NotNil(p.UnitID)
	ts.Equal(unitID, *p.UnitID)
}
func (ts *UseCaseTestSuite) Test_Payload_Alive() {
	var err error
	var payloadStruct *payload.Payload

	err = faker.FakeData(&payloadStruct)
	ts.NoError(err)
	payloadStruct.EndedAt = time.Now().AddDate(0, 0, 1).Unix() // tomorrow

	expired := ts.useCase.IsPayloadExpired(payloadStruct)
	ts.Equal(false, expired)
}

func (ts *UseCaseTestSuite) Test_Payload_Expired() {
	var err error
	var payloadStruct *payload.Payload

	err = faker.FakeData(&payloadStruct)
	ts.NoError(err)
	payloadStruct.EndedAt = time.Now().AddDate(0, 0, -10).Unix() // 10 days ago

	expired := ts.useCase.IsPayloadExpired(payloadStruct)
	ts.Equal(true, expired)
}

func (ts *UseCaseTestSuite) Test_Payload() {
	var err error
	var payloadStruct *payload.Payload

	err = faker.FakeData(&payloadStruct)
	ts.NoError(err)
	payloadString := ts.useCase.BuildPayloadString(payloadStruct)
	resultPayload, err := ts.useCase.ParsePayload(payloadString)

	ts.Equal(payloadStruct.Country, resultPayload.Country)
	ts.Equal(payloadStruct.EndedAt, resultPayload.EndedAt)
	ts.Equal(payloadStruct.Gender, resultPayload.Gender)
	ts.Equal(payloadStruct.YearOfBirth, resultPayload.YearOfBirth)
}

var (
	_ suite.SetupTestSuite = &UseCaseTestSuite{}
)

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseTestSuite))
}

type UseCaseTestSuite struct {
	suite.Suite
	useCase payload.UseCase
}

func (ts *UseCaseTestSuite) SetupTest() {
	ts.useCase = payload.NewUseCase()
}
