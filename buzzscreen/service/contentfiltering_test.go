package service_test

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/service"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbdevice"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	devicerepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/device/repo"
	"github.com/Buzzvil/buzzscreen-api/tests"
)

type (
	// ESContentTestCase type definition
	ESContentTestCase struct {
		name            string
		allocReq        dto.ContentAllocV1Request
		contentCampaign dto.ESContentCampaign
		beforeAction    func() error
		afterAction     func() error
		expected        int
	}
)

func (tc *ESContentTestCase) setContentAllocV1RequestField(field string, value interface{}) {
	v := reflect.ValueOf(&tc.allocReq).Elem().FieldByName(field)
	if v.IsValid() {
		switch value.(type) {
		case int:
			v.SetInt(int64(value.(int)))
		case int64:
			v.SetInt(value.(int64))
		case string:
			v.SetString(value.(string))
		case bool:
			v.SetBool(value.(bool))
		}
	}
}

func newBaseContentAllocV1Request() dto.ContentAllocV1Request {
	return dto.ContentAllocV1Request{
		IFA:             "069e6a97-b341-43a0-b9fc-df2556f06a25",
		AppIDReq:        tests.HsKrAppID,
		UnitDeviceToken: "1b66a654cd0d4c4fb471f4fb02b65015",
		DeviceID:        1092,
		DeviceName:      "",
		DeviceOs:        0,
		YearOfBirthReq:  0,
		Gender:          "",
		Carrier:         "",
		Region:          "",
		SdkVersion:      1280,
		Language:        "",
		Request: &http.Request{
			RemoteAddr: "127.0.0.1:",
		},
	}
}

func newTargetingTestCase(name string, expected int) *ESContentTestCase {
	contentCampaign := createBaseContentCampaign()
	requestParams := newBaseContentAllocV1Request()

	return &ESContentTestCase{
		name:            name,
		allocReq:        requestParams,
		contentCampaign: *contentCampaign,
		expected:        expected,
	}
}

func runTargetingTestCases(t *testing.T, testCases []ESContentTestCase) {
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s", tc.name), func(t *testing.T) {
			insertContentCampaignsToESAndDB(t, &tc.contentCampaign)
			if tc.beforeAction != nil {
				err := tc.beforeAction()
				if err != nil {
					t.Fatalf("Failed to call beforeAction. %v", err)
				}
			}

			esContentCampaigns, err := service.GetV1ContentCampaignsFromES(context.Background(), &tc.allocReq)

			t.Log("TestContentTargeting() - error:", err)

			if err != nil {
				t.Fatal(err, esContentCampaigns)
			}

			actual := len(esContentCampaigns)
			if actual != tc.expected {
				t.Fatalf("name %s expected %d actual %d\nrequest: %+v \ncampaign: %+v", tc.name, tc.expected, actual, tc.allocReq, tc.contentCampaign)
			}

			if tc.afterAction != nil {
				err := tc.afterAction()
				if err != nil {
					t.Fatalf("Failed to call afterAction. %v", err)
				}
			}
			deleteContentCampaignsFromESAndDB(t, &tc.contentCampaign)
		})
	}
}

func TestBasicTargeting(t *testing.T) {
	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	tc = newTargetingTestCase("Basic", 1)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("enabled not match", 0)
	tc.contentCampaign.IsEnabled = false
	testCases = append(testCases, *tc)
	runTargetingTestCases(t, testCases)
}

func TestGenderTargeting(t *testing.T) {
	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	tc = newTargetingTestCase("Gender target mismatch 1", 0)
	tc.contentCampaign.TargetGender = "F"
	tc.setContentAllocV1RequestField("Gender", "")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Gender female target not match", 0)
	tc.contentCampaign.TargetGender = "F"
	tc.setContentAllocV1RequestField("Gender", "M")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Gender female target match", 1)
	tc.contentCampaign.TargetGender = "F"
	tc.setContentAllocV1RequestField("Gender", "F")
	testCases = append(testCases, *tc)

	runTargetingTestCases(t, testCases)
}

func TestAgeTargeting(t *testing.T) {
	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	tc = newTargetingTestCase("Age target mismatch 1", 0)
	tc.contentCampaign.TargetAgeMin = 10
	tc.contentCampaign.TargetAgeMax = 20
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Age target mismatch min", 0)
	tc.contentCampaign.TargetAgeMin = 10
	tc.contentCampaign.TargetAgeMax = 20
	tc.setContentAllocV1RequestField("YearOfBirthReq", yearOfBirth(9))
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Age target mismatch max", 0)
	tc.contentCampaign.TargetAgeMin = 10
	tc.contentCampaign.TargetAgeMax = 20
	tc.setContentAllocV1RequestField("YearOfBirthReq", yearOfBirth(21))
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Age target match lower bound", 1)
	tc.contentCampaign.TargetAgeMin = 10
	tc.contentCampaign.TargetAgeMax = 20
	tc.setContentAllocV1RequestField("YearOfBirthReq", yearOfBirth(10))
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Age target match upper bound", 1)
	tc.contentCampaign.TargetAgeMin = 10
	tc.contentCampaign.TargetAgeMax = 20
	tc.setContentAllocV1RequestField("YearOfBirthReq", yearOfBirth(20))
	testCases = append(testCases, *tc)

	runTargetingTestCases(t, testCases)
}

func TestSdkTargeting(t *testing.T) {
	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	tc = newTargetingTestCase("Sdk target mismatch min", 0)
	tc.contentCampaign.TargetSdkMin = 2000
	tc.setContentAllocV1RequestField("SdkVersion", 1900)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Sdk target mismatch max", 0)
	tc.contentCampaign.TargetSdkMax = 3000
	tc.setContentAllocV1RequestField("SdkVersion", 3100)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Sdk target match lower bound", 1)
	tc.contentCampaign.TargetSdkMin = 2000
	tc.contentCampaign.TargetSdkMax = 3000
	tc.setContentAllocV1RequestField("SdkVersion", 2000)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Sdk target match upper bound", 1)
	tc.contentCampaign.TargetSdkMin = 2000
	tc.contentCampaign.TargetSdkMax = 3000
	tc.setContentAllocV1RequestField("SdkVersion", 3000)
	testCases = append(testCases, *tc)

	runTargetingTestCases(t, testCases)
}

func TestOsVersionTargeting(t *testing.T) {
	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	tc = newTargetingTestCase("OsVersion target mismatch min", 0)
	tc.contentCampaign.TargetOsMin = 22
	tc.setContentAllocV1RequestField("DeviceOs", 21)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("OsVersion target mismatch min", 0)
	tc.contentCampaign.TargetOsMax = 24
	tc.setContentAllocV1RequestField("DeviceOs", 25)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("OsVersion target match lower bound", 1)
	tc.contentCampaign.TargetOsMin = 20
	tc.contentCampaign.TargetOsMax = 26
	tc.setContentAllocV1RequestField("DeviceOs", 20)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("OsVersion target match lower bound", 1)
	tc.contentCampaign.TargetOsMin = 20
	tc.contentCampaign.TargetOsMax = 26
	tc.setContentAllocV1RequestField("DeviceOs", 26)
	testCases = append(testCases, *tc)

	runTargetingTestCases(t, testCases)
}

func TestBatteryOptimizationTargeting(t *testing.T) {
	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	tc = newTargetingTestCase("BatteryOptimization target not match 1", 0)
	tc.contentCampaign.TargetBatteryOptimization = true
	tc.setContentAllocV1RequestField("IsInBatteryOpts", false)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("BatteryOptimization target match 1", 1)
	tc.contentCampaign.TargetBatteryOptimization = true
	tc.setContentAllocV1RequestField("IsInBatteryOpts", true)
	testCases = append(testCases, *tc)

	// test with default value (false)
	tc = newTargetingTestCase("BatteryOptimization target match 2", 1)
	tc.setContentAllocV1RequestField("IsInBatteryOpts", false)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("BatteryOptimization target match 3", 1)
	tc.setContentAllocV1RequestField("IsInBatteryOpts", true)
	testCases = append(testCases, *tc)

	runTargetingTestCases(t, testCases)
}

func TestUnitTargeting(t *testing.T) {
	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	ctx := context.Background()
	appUseCase := buzzscreen.Service.AppUseCase

	tc = newTargetingTestCase("Global target match", 1)
	tc.setContentAllocV1RequestField("AppIDReq", tests.HsKrAppID)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Unit target match 1", 1)
	tc.contentCampaign.TargetAppID = fmt.Sprintf("%d", tests.HsKrAppID)
	hsKrUnit, _ := appUseCase.GetUnitByAppIDAndType(ctx, tests.HsKrAppID, app.UnitTypeLockscreen)
	tc.contentCampaign.TargetOrg = fmt.Sprintf("%d", hsKrUnit.OrganizationID)
	tc.setContentAllocV1RequestField("AppIDReq", tests.HsKrAppID)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Unit target match 2", 1)
	testUnit, _ := appUseCase.GetUnitByAppID(ctx, tests.TestAppID1)
	tc.contentCampaign.TargetUnit = fmt.Sprintf(
		"%d,%d",
		hsKrUnit.ID,
		testUnit.ID,
	)
	tc.setContentAllocV1RequestField("AppIDReq", tests.HsKrAppID)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Unit target match 3", 1)
	tc.contentCampaign.TargetUnit = fmt.Sprintf("%d,%d", hsKrUnit.ID, testUnit.ID)
	tc.setContentAllocV1RequestField("AppIDReq", tests.TestAppID1)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Unit target not match", 0)
	tc.contentCampaign.TargetUnit = fmt.Sprintf("%d,%d", hsKrUnit.ID, testUnit.ID)
	tc.setContentAllocV1RequestField("AppIDReq", tests.TestAppID2)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Unit target unit only match", 1)
	testContentOnlyUnit, _ := appUseCase.GetUnitByAppIDAndType(ctx, tests.TestAppIDContentUnitOnly, app.UnitTypeLockscreen)
	tc.contentCampaign.TargetUnit = fmt.Sprintf("%d,%v", testContentOnlyUnit.ID, model.ESGlobString)
	tc.setContentAllocV1RequestField("AppIDReq", tests.TestAppIDContentUnitOnly)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Unit target unit only not match", 0)
	tc.setContentAllocV1RequestField("AppIDReq", tests.TestAppIDContentUnitOnly)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Unit target unit only not match to targeting global content", 0)

	tc.contentCampaign.TargetUnit = fmt.Sprintf("%d,%v", testContentOnlyUnit.ID+1, model.ESGlobString)
	tc.setContentAllocV1RequestField("AppIDReq", tests.TestAppIDContentUnitOnly)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Unit target unit  match to targeting global content", 1)
	tc.contentCampaign.TargetUnit = fmt.Sprintf("%d,%v", hsKrUnit.ID+1, model.ESGlobString)
	tc.setContentAllocV1RequestField("AppIDReq", tests.HsKrAppID)
	testCases = append(testCases, *tc)

	// Unit de-targeting
	// old unit-detargeting case for checking backward compatibility
	// this test should be deleted after buzzscreen python remove backward compatibility
	tc = newTargetingTestCase("Unit de-target match 1 old", 1)
	tc.contentCampaign.TargetUnit = fmt.Sprintf("%s,-%d", model.ESGlobString, testUnit.ID)
	tc.setContentAllocV1RequestField("AppIDReq", tests.HsKrAppID)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Unit de-target match 1", 1)
	tc.contentCampaign.TargetUnit = fmt.Sprintf("%s", model.ESGlobString)
	tc.contentCampaign.DetargetUnit = fmt.Sprintf("%d", testUnit.ID)
	tc.setContentAllocV1RequestField("AppIDReq", tests.HsKrAppID)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Unit de-target not match 1", 0)
	tc.contentCampaign.TargetUnit = fmt.Sprintf("%s", model.ESGlobString)
	tc.contentCampaign.DetargetUnit = fmt.Sprintf("%d", testUnit.ID)
	tc.setContentAllocV1RequestField("AppIDReq", tests.TestAppID1)
	testCases = append(testCases, *tc)

	runTargetingTestCases(t, testCases)
}

func TestOrganizationTargeting(t *testing.T) {
	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	ctx := context.Background()
	appUseCase := buzzscreen.Service.AppUseCase

	tc = newTargetingTestCase("Organization target match 1", 1)
	hsKrUnit, _ := appUseCase.GetUnitByAppIDAndType(ctx, tests.HsKrAppID, app.UnitTypeLockscreen)
	testUnit1, _ := appUseCase.GetUnitByAppID(ctx, tests.TestAppID1)
	testUnit2, _ := appUseCase.GetUnitByAppID(ctx, tests.TestAppID2)
	tc.contentCampaign.TargetOrg = fmt.Sprintf(
		"%d,%d",
		hsKrUnit.OrganizationID,
		hsKrUnit.OrganizationID+1,
	)
	tc.setContentAllocV1RequestField("AppIDReq", tests.HsKrAppID)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Organization target match 2", 1)
	tc.contentCampaign.TargetOrg = fmt.Sprintf(
		"%d,%d",
		testUnit1.OrganizationID+1,
		testUnit1.OrganizationID,
	)
	tc.setContentAllocV1RequestField("AppIDReq", tests.TestAppID1)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Organization target not match", 0)
	tc.contentCampaign.TargetOrg = fmt.Sprintf(
		"%d,%d",
		testUnit2.OrganizationID+1,
		testUnit2.OrganizationID+2,
	)
	tc.setContentAllocV1RequestField("AppIDReq", tests.TestAppID2)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Organization target unit  match to targeting global content", 1)
	tc.contentCampaign.TargetOrg = fmt.Sprintf(
		"%d,%v",
		hsKrUnit.OrganizationID+1,
		model.ESGlobString,
	)
	tc.setContentAllocV1RequestField("AppIDReq", tests.HsKrAppID)
	testCases = append(testCases, *tc)

	// Organization de-targeting

	tc = newTargetingTestCase("Organization de-target match 1", 1)
	tc.contentCampaign.TargetOrg = fmt.Sprintf("%s", model.ESGlobString)
	tc.contentCampaign.DetargetOrg = fmt.Sprintf("%d", hsKrUnit.OrganizationID+1)
	tc.setContentAllocV1RequestField("AppIDReq", tests.HsKrAppID)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Organization de-target not match 1", 0)
	tc.contentCampaign.TargetOrg = fmt.Sprintf("%s", model.ESGlobString)
	tc.contentCampaign.DetargetOrg = fmt.Sprintf("%d", testUnit1.OrganizationID)
	tc.setContentAllocV1RequestField("AppIDReq", tests.TestAppID1)
	testCases = append(testCases, *tc)

	runTargetingTestCases(t, testCases)
}

func TestAppIDTargeting(t *testing.T) {
	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	ctx := context.Background()
	appUseCase := buzzscreen.Service.AppUseCase

	tc = newTargetingTestCase("App ID target match 1", 1)
	hsKrUnit, _ := appUseCase.GetUnitByAppIDAndType(ctx, tests.HsKrAppID, app.UnitTypeLockscreen)
	tc.contentCampaign.TargetOrg = fmt.Sprintf("%d", hsKrUnit.OrganizationID)
	tc.setContentAllocV1RequestField("AppIDReq", tests.HsKrAppID)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("App ID target match 2", 1)
	tc.contentCampaign.TargetAppID = fmt.Sprintf("%d,%d", tests.HsKrAppID, tests.TestAppID1)
	tc.setContentAllocV1RequestField("AppIDReq", tests.HsKrAppID)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("App ID target match 3", 1)
	tc.contentCampaign.TargetAppID = fmt.Sprintf("%d,%d", tests.HsKrAppID, tests.TestAppID1)
	tc.setContentAllocV1RequestField("AppIDReq", tests.TestAppID1)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("App ID target not match", 0)
	tc.contentCampaign.TargetAppID = fmt.Sprintf("%d,%d", tests.HsKrAppID, tests.TestAppID1)
	tc.setContentAllocV1RequestField("AppIDReq", tests.TestAppID2)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("App ID target app id  match to targeting global content", 1)
	tc.contentCampaign.TargetAppID = fmt.Sprintf("%d,%v", tests.HsKrAppID+1, model.ESGlobString)
	tc.setContentAllocV1RequestField("AppIDReq", tests.HsKrAppID)
	testCases = append(testCases, *tc)

	// App ID de-targeting
	tc = newTargetingTestCase("App ID de-target match 1", 1)
	tc.contentCampaign.TargetAppID = fmt.Sprintf("%s", model.ESGlobString)
	tc.contentCampaign.DetargetAppID = fmt.Sprintf("%d", tests.TestAppID1)
	tc.setContentAllocV1RequestField("AppIDReq", tests.HsKrAppID)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Unit de-target not match 1", 0)
	tc.contentCampaign.TargetAppID = fmt.Sprintf("%s", model.ESGlobString)
	tc.contentCampaign.DetargetAppID = fmt.Sprintf("%d", tests.TestAppID1)
	tc.setContentAllocV1RequestField("AppIDReq", tests.TestAppID1)
	testCases = append(testCases, *tc)

	runTargetingTestCases(t, testCases)
}

func TestDateTargeting(t *testing.T) {
	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	// start_date/end_date targeting
	// elasticsearch 쿼리를 시간단위로 타게팅되도록 했기 때문에 EndDate에 2 hour를 더해줘야 한다
	// 1 hour가 아닌 이유는 2시 59분 59초에 index가 된 경우 할당시 1시간 밀려서 타게팅 안되는 경우가 발생할 수 있기 때문
	tc = newTargetingTestCase("date match", 1)
	tc.contentCampaign.StartDate = time.Now().Add(-1 * time.Hour).Format("2006-01-02T15:04:05")
	tc.contentCampaign.EndDate = time.Now().Add(2 * time.Hour).Format("2006-01-02T15:04:05")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("past date not match", 0)
	tc.contentCampaign.StartDate = time.Now().Add(-2 * time.Hour).Format("2006-01-02T15:04:05")
	tc.contentCampaign.EndDate = time.Now().Add(-1 * time.Hour).Format("2006-01-02T15:04:05")
	testCases = append(testCases, *tc)

	// 3시간 전에 미리 할당하는데 마찬 가지로 시간의 경계에 테스트가 돈 경우 1시간 오차가 발생해서 4시간 후로 세팅해놓은 컨텐츠가 할당 될 수 있다
	// 안전하게 5시간 이후로 세팅
	tc = newTargetingTestCase("future date not match", 0)
	tc.contentCampaign.StartDate = time.Now().Add(5 * time.Hour).Format("2006-01-02T15:04:05")
	tc.contentCampaign.EndDate = time.Now().Add(10 * time.Hour).Format("2006-01-02T15:04:05")
	testCases = append(testCases, *tc)

	runTargetingTestCases(t, testCases)
}

func TestCarrierTargeting(t *testing.T) {
	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	tc = newTargetingTestCase("empty carrier not match", 0)
	tc.contentCampaign.TargetCarrier = "abc,def"
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("carrier not match", 0)
	tc.contentCampaign.TargetCarrier = "abc,def"
	tc.setContentAllocV1RequestField("Carrier", "aaa")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("carrier match", 1)
	tc.contentCampaign.TargetCarrier = "abc,def"
	tc.setContentAllocV1RequestField("Carrier", "def")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("carrier match", 1)
	tc.contentCampaign.TargetCarrier = "skt"
	tc.setContentAllocV1RequestField("Carrier", "SKTelecom")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("carrier match", 1)
	tc.contentCampaign.TargetCarrier = "skt,kt"
	tc.setContentAllocV1RequestField("Carrier", "3G olleh")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("carrier match", 1)
	tc.contentCampaign.TargetCarrier = "twm"
	tc.setContentAllocV1RequestField("Carrier", "台灣大哥大")
	testCases = append(testCases, *tc)

	runTargetingTestCases(t, testCases)
}

func TestContentFilterProvider(t *testing.T) {
	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	tc = newTargetingTestCase("DetargetingProvider", 0)
	tc.contentCampaign.ProviderID = tests.FilteringProviderID
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Ok", 1)
	tc.contentCampaign.ProviderID = tests.FilteringProviderID + 1
	testCases = append(testCases, *tc)

	runTargetingTestCases(t, testCases)
}

func TestCountryTargeting(t *testing.T) {
	//TODO: country는 값 없으면 타게팅 안되게 막자 - zune
	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	tc = newTargetingTestCase("Use country param rather than unit's country", 1)
	tc.contentCampaign.Country = "US"
	tc.setContentAllocV1RequestField("AppIDReq", tests.KoreaUnit.AppID)
	tc.setContentAllocV1RequestField("CountryReq", "US")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("country not match", 0)
	tc.contentCampaign.Country = "US"
	tc.setContentAllocV1RequestField("AppIDReq", tests.KoreaUnit.AppID)
	if tc.contentCampaign.Country == tests.KoreaUnit.Country {
		t.Fatalf("Country should be different")
	}
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Global unit with no country", 0)
	tc.contentCampaign.Country = "KR"
	tc.setContentAllocV1RequestField("AppIDReq", tests.GlobalUnit.AppID)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Global unit with no country match", 0)
	tc.contentCampaign.Country = "KR"
	tc.setContentAllocV1RequestField("AppIDReq", tests.GlobalUnit.AppID)
	tc.setContentAllocV1RequestField("CountryReq", "JP")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Global unit with country match", 1)
	tc.contentCampaign.Country = "US"
	tc.setContentAllocV1RequestField("AppIDReq", tests.GlobalUnit.AppID)
	tc.setContentAllocV1RequestField("CountryReq", "US")
	testCases = append(testCases, *tc)

	runTargetingTestCases(t, testCases)
}

func TestLanguageTargeting(t *testing.T) {
	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	tc = newTargetingTestCase("Matching languages", 1)
	tc.contentCampaign.TargetLanguage = "elvish"
	tc.setContentAllocV1RequestField("Language", "elvish")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Not maching languages", 0)
	tc.contentCampaign.TargetLanguage = "ko"
	tc.setContentAllocV1RequestField("Language", "ja")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Request without language", 1)
	tc.contentCampaign.TargetLanguage = "ko"
	tc.setContentAllocV1RequestField("Language", "")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Content without language", 0)
	tc.contentCampaign.TargetLanguage = ""
	tc.setContentAllocV1RequestField("Language", "ko")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("Content and request without language", 1)
	tc.contentCampaign.TargetLanguage = ""
	tc.setContentAllocV1RequestField("Language", "")
	testCases = append(testCases, *tc)

	runTargetingTestCases(t, testCases)
}

func TestCustomTargeting(t *testing.T) {
	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	nums := []string{"1", "2", "3"}
	for _, num := range nums {
		tc = newTargetingTestCase(fmt.Sprintf("custom %s not match with empty request", num), 0)
		reflect.ValueOf(&tc.contentCampaign).Elem().
			FieldByName(fmt.Sprintf("CustomTarget%s", num)).
			SetString("abc,def")
		tc.setContentAllocV1RequestField(fmt.Sprintf("CustomTarget%s", num), "")
		testCases = append(testCases, *tc)

		tc = newTargetingTestCase(fmt.Sprintf("custom %s not match ", num), 0)
		reflect.ValueOf(&tc.contentCampaign).Elem().
			FieldByName(fmt.Sprintf("CustomTarget%s", num)).
			SetString("abc,def")
		tc.setContentAllocV1RequestField(fmt.Sprintf("CustomTarget%s", num), "aaa,bbb")
		testCases = append(testCases, *tc)

		tc = newTargetingTestCase(fmt.Sprintf("custom %s match 1", num), 1)
		reflect.ValueOf(&tc.contentCampaign).Elem().
			FieldByName(fmt.Sprintf("CustomTarget%s", num)).
			SetString("abc,def")
		tc.setContentAllocV1RequestField(fmt.Sprintf("CustomTarget%s", num), "def,cc")
		testCases = append(testCases, *tc)

		tc = newTargetingTestCase(fmt.Sprintf("custom %s match 2", num), 1)
		reflect.ValueOf(&tc.contentCampaign).Elem().
			FieldByName(fmt.Sprintf("CustomTarget%s", num)).
			SetString("abc,def")
		tc.setContentAllocV1RequestField(fmt.Sprintf("CustomTarget%s", num), "def,abc")
		testCases = append(testCases, *tc)
	}

	runTargetingTestCases(t, testCases)
}

func TestRegionTargeting(t *testing.T) {
	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	tc = newTargetingTestCase("region target with empty request param", 0)
	tc.contentCampaign.TargetRegion = "서울시 관악구,경상남도"
	tc.setContentAllocV1RequestField("Region", "")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("region target not match 1", 0)
	tc.contentCampaign.TargetRegion = "서울시 관악구,경상남도"
	tc.setContentAllocV1RequestField("Region", "경기도")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("region target not match 2", 0)
	tc.contentCampaign.TargetRegion = "서울시 관악구,경상남도"
	tc.setContentAllocV1RequestField("Region", "경기도 수원시")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("sub region target not match 1", 0)
	tc.contentCampaign.TargetRegion = "서울시 관악구,경상남도"
	tc.setContentAllocV1RequestField("Region", "서울시")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("sub region target not match 2", 0)
	tc.contentCampaign.TargetRegion = "서울시 관악구,경상남도"
	tc.setContentAllocV1RequestField("Region", "서울시 서초구")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("region target match 1", 1)
	tc.contentCampaign.TargetRegion = "서울시 관악구,경상남도"
	tc.setContentAllocV1RequestField("Region", "경상남도")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("region target match 2", 1)
	tc.contentCampaign.TargetRegion = "서울시 관악구,경상남도"
	tc.setContentAllocV1RequestField("Region", "경상남도 거창군")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("sub region target match", 1)
	tc.contentCampaign.TargetRegion = "서울시 관악구,경상남도"
	tc.setContentAllocV1RequestField("Region", "서울시 관악구")
	testCases = append(testCases, *tc)

	runTargetingTestCases(t, testCases)
}

func getWeekdayStartsFromMonday(t *time.Time) int {
	return (int(t.Weekday()) + 6) % 7
}

func timeIn(timezone string) time.Time {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		panic(err)
	}
	return time.Now().In(loc)
}

func TestWeekSlotTargeting(t *testing.T) {
	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	tc = newTargetingTestCase("null week slot target", 1)
	tc.setContentAllocV1RequestField("AppIDReq", tests.KoreaUnit.AppID)
	testCases = append(testCases, *tc)

	localTime := timeIn(tests.KoreaUnit.Timezone)
	weekday := getWeekdayStartsFromMonday(&localTime)
	currentSlot := weekday*24 + localTime.Hour()
	numSlots := 24 * 7
	nextSlot := (currentSlot + 1) % numSlots

	// 현재 slot이 26 이면 week_slot 0 ~ 25, 27 ~ (numSlot - 1)에 대해서 할당이 안되어야 하지만
	// 테스트중에 시간이 지나 현재 slot이 27이 되면 할당이 될 수도 있으므로
	// 0 ~ 25, 28 ~ (numSlot - 1)에 대해서 할당이 안되는지 체크
	weekSlot := make([]string, numSlots)
	for i := 0; i <= numSlots; i++ {
		if i == currentSlot || i == nextSlot {
			continue
		}
		weekSlot = append(weekSlot, strconv.Itoa(i))
	}

	tc = newTargetingTestCase("week slot not match", 0)
	tc.contentCampaign.WeekSlot = strings.Join(weekSlot, ",")
	tc.setContentAllocV1RequestField("AppIDReq", tests.KoreaUnit.AppID)
	testCases = append(testCases, *tc)

	// 마찬가지로 안전하게 현재 slot이 26이면 week_slot을 26,27로 세팅해서 할당 체크
	tc = newTargetingTestCase("week slot match", 1)
	tc.contentCampaign.WeekSlot = fmt.Sprintf("%d,%d", currentSlot, nextSlot)
	tc.setContentAllocV1RequestField("AppIDReq", tests.KoreaUnit.AppID)
	testCases = append(testCases, *tc)

	runTargetingTestCases(t, testCases)
}

func TestCategoryFiltering(t *testing.T) {
	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	tc = newTargetingTestCase("no category request param", 1)
	tc.contentCampaign.Categories = "abc,def"
	tc.setContentAllocV1RequestField("FilterCategories", "")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("category not filtered", 1)
	tc.contentCampaign.Categories = "abc,def"
	tc.setContentAllocV1RequestField("FilterCategories", "aaa,bbb,ccc")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("category filtered", 0)
	tc.contentCampaign.Categories = "abc,def"
	tc.setContentAllocV1RequestField("FilterCategories", "def,ggg")
	testCases = append(testCases, *tc)

	runTargetingTestCases(t, testCases)
}

func TestCategoryTargeting(t *testing.T) {
	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	tc = newTargetingTestCase("no category request param", 1)
	tc.contentCampaign.Categories = "abc,def"
	tc.setContentAllocV1RequestField("Categories", "")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("category not targeted", 0)
	tc.contentCampaign.Categories = "abc,def"
	tc.setContentAllocV1RequestField("Categories", "aaa,bbb,ccc")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("category targeted", 1)
	tc.contentCampaign.Categories = "abc,def"
	tc.setContentAllocV1RequestField("Categories", "def,ggg")
	testCases = append(testCases, *tc)

	runTargetingTestCases(t, testCases)
}

func TestChannelFiltering(t *testing.T) {
	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	var channelID3 int64 = 3
	var channelID2 int64 = 2

	tc = newTargetingTestCase("no channel request param", 1)
	tc.contentCampaign.ChannelID = &channelID3
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("channel not filtered", 1)
	tc.contentCampaign.ChannelID = &channelID3
	tc.setContentAllocV1RequestField("FilterChannelIDs", "1,2")
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("channel filtered", 0)
	tc.contentCampaign.ChannelID = &channelID2
	tc.setContentAllocV1RequestField("FilterChannelIDs", "1,2")
	testCases = append(testCases, *tc)

	runTargetingTestCases(t, testCases)
}

func TestFrequencyCapping(t *testing.T) {
	func() {
		dyDB := env.GetDynamoDB()
		err := dyDB.CreateTable(env.Config.DynamoTableActivity, devicerepo.Activity{}).Run()
		if err != nil && !strings.Contains(err.Error(), "ResourceInUseException") {
			core.Logger.Fatalf("--- Failed. SetupTest failed with %v", err)
		}

		dpTable := dyDB.Table(env.Config.DynamoTableProfile)
		dpr := devicerepo.NewProfileRepo(&dpTable)
		daTable := dyDB.Table(env.Config.DynamoTableActivity)
		dar := devicerepo.NewActivityRepo(&daTable)
		dr := devicerepo.New(dbdevice.NewSource(buzzscreen.Service.DB))
		buzzscreen.Service.DeviceUseCase = device.NewUseCase(dr, dpr, dar)
	}()

	saveActivityAction := func(deviceID int64, campaignID int64, now time.Time, agos ...time.Duration) func() error {
		return func() error {
			dyDB := env.GetDynamoDB()
			err := dyDB.CreateTable(env.Config.DynamoTableActivity, devicerepo.Activity{}).Run()
			if err != nil && !strings.Contains(err.Error(), "ResourceInUseException") {
				return err
			}

			table := dyDB.Table(env.Config.DynamoTableActivity)
			for _, ago := range agos {
				a := devicerepo.Activity{
					DeviceID:   deviceID,
					CreatedAt:  float64(time.Now().Add(-ago).Unix()),
					ActionType: "i",
					CampaignID: campaignID,
					TTL:        time.Now().AddDate(0, 0, 2).Unix(),
				}

				err := table.Put(&a).Run()
				if err != nil {
					return err
				}
			}
			return nil
		}
	}

	deleteActivitiesAction := func() func() error {
		return func() error {
			return env.GetDynamoDB().Table(env.Config.DynamoTableActivity).DeleteTable().Run()
		}
	}

	getActions := func(deviceID int64, campaginID int64, agos ...time.Duration) (func() error, func() error) {
		now := time.Now()
		return saveActivityAction(deviceID, campaginID, now, agos...), deleteActivitiesAction()
	}
	timeDay := time.Hour * 24

	testCases := make([]ESContentTestCase, 0, 16)
	var tc *ESContentTestCase

	tc = newTargetingTestCase("frequency capping match", 1)
	tc.beforeAction, tc.afterAction = getActions(
		tc.allocReq.DeviceID,
		tc.contentCampaign.ID,
		time.Minute, time.Minute*2, time.Minute*59, time.Minute*61,
	)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("frequency capping match ipu max", 1)
	tc.beforeAction, tc.afterAction = getActions(
		tc.allocReq.DeviceID,
		tc.contentCampaign.ID,
		time.Minute, time.Minute*2, time.Minute*58, time.Minute*59, time.Minute*61,
	)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("frequency capping mismatch ipu min", 0)
	tc.beforeAction, tc.afterAction = getActions(
		tc.allocReq.DeviceID,
		tc.contentCampaign.ID,
		time.Minute, time.Minute*2, time.Minute*57, time.Minute*58, time.Minute*59, time.Minute*61,
	)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("frequency capping match dipu max", 1)
	tc.beforeAction, tc.afterAction = getActions(
		tc.allocReq.DeviceID,
		tc.contentCampaign.ID,
		time.Minute, time.Minute*2, time.Minute*58, time.Minute*59, time.Minute*61,
		timeDay-time.Minute*4, timeDay-time.Minute*3, timeDay-time.Minute*2, timeDay-time.Minute, timeDay+time.Minute,
	)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("frequency capping mismatch dipu min", 0)
	tc.beforeAction, tc.afterAction = getActions(
		tc.allocReq.DeviceID,
		tc.contentCampaign.ID,
		time.Minute, time.Minute*2, time.Minute*58, time.Minute*59, time.Minute*61,
		timeDay-time.Minute*5, timeDay-time.Minute*4, timeDay-time.Minute*3, timeDay-time.Minute*2, timeDay-time.Minute, timeDay+time.Minute,
	)
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("frequency capping match nil ipu/dipu with no activities", 1)
	tc.contentCampaign.Ipu = nil
	tc.contentCampaign.Dipu = nil
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("frequency capping match nil ipu/dipu with activities", 1)
	tc.beforeAction, tc.afterAction = getActions(
		tc.allocReq.DeviceID,
		tc.contentCampaign.ID,
		time.Minute, time.Minute*2, time.Minute*58, time.Minute*59, time.Minute*61,
		timeDay-time.Minute*5, timeDay-time.Minute*4, timeDay-time.Minute*3, timeDay-time.Minute*2, timeDay-time.Minute, timeDay+time.Minute,
	)
	tc.contentCampaign.Ipu = nil
	tc.contentCampaign.Dipu = nil
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("frequency capping mismatch dipu min with nil ipu ", 0)
	tc.beforeAction, tc.afterAction = getActions(
		tc.allocReq.DeviceID,
		tc.contentCampaign.ID,
		time.Minute, time.Minute*2, time.Minute*58, time.Minute*59, time.Minute*61,
		timeDay-time.Minute*5, timeDay-time.Minute*4, timeDay-time.Minute*3, timeDay-time.Minute*2, timeDay-time.Minute, timeDay+time.Minute,
	)
	tc.contentCampaign.Ipu = nil
	testCases = append(testCases, *tc)

	tc = newTargetingTestCase("frequency capping mismatch ipu min with nil dipu ", 0)
	tc.beforeAction, tc.afterAction = getActions(
		tc.allocReq.DeviceID,
		tc.contentCampaign.ID,
		time.Minute, time.Minute*2, time.Minute*57, time.Minute*58, time.Minute*59, time.Minute*61,
	)
	tc.contentCampaign.Dipu = nil
	testCases = append(testCases, *tc)

	runTargetingTestCases(t, testCases)
}
