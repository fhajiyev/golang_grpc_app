package event

import (
	"bytes"
	"encoding/gob"
	"math/rand"
	"testing"

	"github.com/Buzzvil/buzzlib-go/header"
	"github.com/bxcodec/faker"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseTestSuite))
}

func (ts *UseCaseTestSuite) TestSaveTrackURL() {
	deviceID := rand.Int63n(1000000) + 1
	resource := Resource{
		ID:   rand.Int63n(1000000) + 1,
		Type: ResourceTypeAd,
	}
	trackingURL := "https://tracking.url"

	logMethod := "save"
	logValidator := ts.validateLogTrackURLActivity(logMethod, deviceID, resource, trackingURL)
	ts.structuredLogger.On("Log", mock.MatchedBy(logValidator)).Once()

	ts.repo.On("SaveTrackingURL", deviceID, resource, trackingURL).Once()

	ts.useCase.SaveTrackingURL(deviceID, resource, trackingURL)
}

func (ts *UseCaseTestSuite) TestGetTrackURL() {
	deviceID := rand.Int63n(1000000) + 1
	resource := Resource{
		ID:   rand.Int63n(1000000) + 1,
		Type: ResourceTypeAd,
	}
	trackingURL := "https://tracking.url"

	logMethod := "get"
	logValidator := ts.validateLogTrackURLActivity(logMethod, deviceID, resource, trackingURL)
	ts.structuredLogger.On("Log", mock.MatchedBy(logValidator)).Once()

	ts.repo.On("GetTrackingURL", deviceID, resource).Return(trackingURL, nil).Once()
	ts.repo.On("DeleteTrackingURL", deviceID, resource).Return(nil).Once()

	expected, err := ts.useCase.GetTrackingURL(deviceID, resource)

	ts.NoError(err)
	ts.Equal(trackingURL, expected)
}

func (ts *UseCaseTestSuite) validateLogTrackURLActivity(method string, deviceID int64, resource Resource, trackURL string) func(map[string]interface{}) bool {
	return func(m map[string]interface{}) bool {
		ts.Equal(method, m["method"])
		ts.Equal(deviceID, m["device_id"])
		ts.Equal(resource.ID, m["resource_id"])
		ts.Equal(resource.Type, m["resource_type"])
		ts.Equal(trackURL, m["tracking_url"])
		return true
	}
}

func (ts *UseCaseTestSuite) Test_BuildTokenString() {
	token := ts.createToken()
	seralized := ts.serialize(*token)
	tokenStr := "TEST_TOKEN_STRING"

	te := tokenEncrypter{
		manager: ts.manager,
	}

	ts.manager.On("GenerateToken", seralized).Return(tokenStr, nil).Once()

	resultTokenStr, err := te.Build(*token)
	ts.NoError(err)
	ts.Equal(tokenStr, resultTokenStr)
}

func (ts *UseCaseTestSuite) Test_ParseTokenString() {
	token := ts.createToken()
	seralized := ts.serialize(*token)
	tokenStr := "TEST_TOKEN_STRING"

	te := tokenEncrypter{
		manager: ts.manager,
	}

	ts.manager.On("GetDataByToken", tokenStr).Return(seralized, nil).Once()

	resultToken, err := te.Parse(tokenStr)
	ts.NoError(err)
	ts.Equal(token, resultToken)
}

func (ts *UseCaseTestSuite) Test_TrackEvent() {
	token := ts.createToken()

	ts.handler.On("Publish", *token).Return(nil).Once()

	err := ts.useCase.TrackEvent(ts.handler, *token)
	ts.NoError(err)
	ts.handler.AssertExpectations(ts.T())
}

func (ts *UseCaseTestSuite) Test_GetRewardStatus() {
	token := ts.createToken()
	auth := ts.createAuth()

	status := "pending"
	ts.repo.On("GetRewardStatus", *token, auth).Return(status, nil).Once()

	res, err := ts.useCase.GetRewardStatus(*token, auth)
	ts.NoError(err)
	ts.Equal(status, res)
}

func (ts *UseCaseTestSuite) Test_GetEventsMap() {
	unitID := rand.Int63n(1000000) + 1
	a := ts.createAuth()

	resources := make([]Resource, 0)
	resource := Resource{}
	ts.NoError(faker.FakeData(&resource))
	resources = append(resources, resource)

	eventsMap := make(map[int64]Events)
	events := Events{
		ts.createEvent(TypeClicked),
		ts.createEvent(TypeImpressed),
	}
	eventsMap[resource.ID] = events

	matchTokenEncrypter := mock.MatchedBy(func(te interface{}) bool {
		_, ok := te.(TokenEncrypter)
		return ok
	})
	ts.repo.On("GetEventsMap", resources, unitID, a, matchTokenEncrypter).Return(eventsMap, nil).Once()

	resultEventsMap, err := ts.useCase.GetEventsMap(resources, unitID, a)
	ts.NoError(err)

	resultEvents, ok := resultEventsMap[resource.ID]
	ts.True(ok)
	ts.Equal(events, resultEvents)
	ts.repo.AssertExpectations(ts.T())

}

func (ts *UseCaseTestSuite) createToken() *Token {
	return &Token{
		TransactionID: uuid.NewV4().String(),
		Resource:      ts.createResource(),
		EventType:     "landed",
		UnitID:        rand.Int63n(100000) + 1,
	}
}

func (ts *UseCaseTestSuite) createResource() Resource {
	return Resource{
		ID:   rand.Int63n(100000) + 1,
		Type: ResourceTypeAd,
	}
}

func (ts *UseCaseTestSuite) createAuth() header.Auth {
	a := header.Auth{}
	ts.NoError(faker.FakeData(&a))
	return a
}

func (ts *UseCaseTestSuite) createEvent(eventType string) Event {
	r := &Reward{}
	ts.NoError(faker.FakeData(&r))
	return Event{
		Type:         eventType,
		TrackingURLs: []string{"trackingurl1", "trackingurl2"},
		Reward:       r,
	}
}

func (ts *UseCaseTestSuite) serialize(token Token) []byte {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)

	err := e.Encode(token)
	ts.NoError(err)

	return b.Bytes()
}

func (ts *UseCaseTestSuite) deserialize(serialized []byte) *Token {
	b := bytes.Buffer{}
	b.Write(serialized)
	d := gob.NewDecoder(&b)

	token := Token{}
	err := d.Decode(&token)
	ts.NoError(err)

	return &token
}

var (
	_ suite.SetupTestSuite = &UseCaseTestSuite{}
)

type UseCaseTestSuite struct {
	suite.Suite
	repo             *mockRepo
	handler          *mockMessageHandler
	manager          *mockJWEManager
	structuredLogger *mockStructuredLogger
	useCase          UseCase
}

func (ts *UseCaseTestSuite) SetupTest() {
	ts.repo = new(mockRepo)
	ts.handler = new(mockMessageHandler)
	ts.manager = new(mockJWEManager)
	ts.structuredLogger = new(mockStructuredLogger)
	ts.useCase = NewUseCase(ts.repo, ts.manager, ts.structuredLogger)
}

func (ts *UseCaseTestSuite) AfterTest(_, _ string) {
	ts.repo.AssertExpectations(ts.T())
	ts.handler.AssertExpectations(ts.T())
	ts.manager.AssertExpectations(ts.T())
	ts.structuredLogger.AssertExpectations(ts.T())
}

type mockRepo struct {
	mock.Mock
}

func (r *mockRepo) GetRewardStatus(token Token, auth header.Auth) (string, error) {
	ret := r.Called(token, auth)
	return ret.Get(0).(string), ret.Error(1)
}

func (r *mockRepo) GetEventsMap(resources []Resource, unitID int64, auth header.Auth, tokenEncrypter TokenEncrypter) (map[int64]Events, error) {
	ret := r.Called(resources, unitID, auth, tokenEncrypter)
	return ret.Get(0).(map[int64]Events), ret.Error(1)
}

func (r *mockRepo) ParseToken(token string) (*Token, error) {
	ret := r.Called(token)
	return ret.Get(0).(*Token), ret.Error(1)
}

func (r *mockRepo) SaveTrackingURL(deviceID int64, resource Resource, trackURL string) {
	r.Called(deviceID, resource, trackURL)
}

func (r *mockRepo) GetTrackingURL(deviceID int64, resource Resource) (string, error) {
	ret := r.Called(deviceID, resource)
	return ret.Get(0).(string), ret.Error(1)
}

func (r *mockRepo) DeleteTrackingURL(deviceID int64, resource Resource) error {
	ret := r.Called(deviceID, resource)
	return ret.Error(0)
}

type mockMessageHandler struct {
	mock.Mock
}

func (r *mockMessageHandler) Publish(ingredients Token) error {
	ret := r.Called(ingredients)
	return ret.Error(0)
}

type mockJWEManager struct {
	mock.Mock
}

func (m *mockJWEManager) GenerateToken(marshaled []byte) (string, error) {
	args := m.Called(marshaled)
	return args.Get(0).(string), args.Error(1)
}

func (m *mockJWEManager) GetDataByToken(token string) ([]byte, error) {
	args := m.Called(token)
	return args.Get(0).([]byte), args.Error(1)
}

type mockStructuredLogger struct {
	mock.Mock
}

func (l *mockStructuredLogger) Log(m map[string]interface{}) {
	l.Called(m)
}
