package location_test

import (
	"net"
	"net/http"
	"testing"

	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/mock"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/location"
	"github.com/stretchr/testify/suite"
)

func (ts *UseCaseTestSuite) Test_GetClientLocation() {
	mockLocation := &location.Location{
		Country:   "US",
		ZipCode:   "55454",
		State:     "MN",
		City:      "Minneapolis",
		TimeZone:  "CST",
		Latitude:  94.14,
		Longitude: -87.24,
		IPAddress: "127.0.0.1",
	}
	ts.repo.On("GetClientLocation", mock.Anything, mock.Anything).Return(mockLocation, nil).Once()

	result := ts.useCase.GetClientLocation(new(http.Request), "USA")

	ts.Equal(mockLocation.Country, result.Country)
	ts.repo.AssertExpectations(ts.T())
}

func (ts *UseCaseTestSuite) Test_GetCountryFromIP() {
	var ip net.IP
	ts.NoError(faker.FakeData(&ip))
	cou := "KR"
	ts.repo.On("GetCountryFromIP", ip).Return(cou, nil).Once()

	res, err := ts.useCase.GetCountryFromIP(ip)

	ts.NoError(err)
	ts.Equal(cou, res)
	ts.repo.AssertExpectations(ts.T())
}

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseTestSuite))
}

var (
	_ suite.SetupTestSuite = &UseCaseTestSuite{}
)

type UseCaseTestSuite struct {
	suite.Suite
	repo    *mockRepo
	useCase location.UseCase
}

func (ts *UseCaseTestSuite) SetupTest() {
	ts.repo = new(mockRepo)
	ts.useCase = location.NewUseCase(ts.repo)
}

var _ location.Repository = &mockRepo{}

type mockRepo struct {
	mock.Mock
}

func (r *mockRepo) GetClientLocation(httpRequest *http.Request, countryFromLocale string) *location.Location {
	ret := r.Called(httpRequest, countryFromLocale)
	return ret.Get(0).(*location.Location)
}

func (r *mockRepo) GetCountryFromIP(ip net.IP) (string, error) {
	ret := r.Called(ip)
	return ret.Get(0).(string), ret.Error(1)
}
