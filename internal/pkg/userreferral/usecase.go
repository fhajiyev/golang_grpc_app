package userreferral

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/kenshaw/baseconv"
)

const (
	codeEcodingChars = "345689bcdghijkmnopqrstuvwxyz" // 25 chars, we can support 25^8 without issue
	codePaddingChars = "a4e27f"                       // random 7 character for padding NOTE they shouldn't belong in encodingChars
	codeMinDigit     = 8
)

// NewUseCase returns UseCase interface
func NewUseCase(userRepo Repository) UseCase {
	return &useCase{userRepo: userRepo}
}

// UseCase userreferral
type UseCase interface {
	GetUserByCode(code string) (*DeviceUser, error)
	GetOrCreateUserByDevice(deviceID int64, appID int64, udt string, verifyURL string) (*DeviceUser, error)
	CreateReferral(ingr CreateReferralIngredients) (bool, error)
}

type useCase struct {
	userRepo Repository
}

// GetUserByCode func definition
func (u *useCase) GetUserByCode(code string) (*DeviceUser, error) {
	if code == "" {
		return nil, InvalidArgumentError{ArgName: "code", ArgValue: code}
	}
	return u.userRepo.GetUserByCode(code)
}

// GetOrCreateUserByDevice func definition
func (u *useCase) GetOrCreateUserByDevice(deviceID int64, appID int64, udt string, verifyURL string) (*DeviceUser, error) {
	if deviceID == 0 {
		return nil, InvalidArgumentError{ArgName: "deviceID", ArgValue: 0}
	}

	// Whether device exists or not should be checked in controller using device domain
	user, err := u.userRepo.GetUserByDevice(deviceID)
	_, notfound := err.(NotFoundError)
	if notfound {
		return u.createUserForDevice(deviceID, appID, udt, verifyURL)
	} else if err != nil {
		return nil, err
	}

	return user, nil
}
func (u *useCase) CreateReferral(ingr CreateReferralIngredients) (bool, error) {
	if ingr.Code == "" {
		return false, InvalidArgumentError{ArgName: "code", ArgValue: ingr.Code}
	}

	referee, err := u.GetOrCreateUserByDevice(ingr.DeviceID, ingr.AppID, ingr.UnitDeviceToken, ingr.VerifyURL)
	if err != nil {
		return false, err
	}
	referrer, err := u.userRepo.GetUserByCode(ingr.Code)
	_, notfound := err.(NotFoundError)
	if notfound {
		return false, InvalidArgumentError{ArgName: "Code", ArgValue: ingr.Code}
	} else if err != nil {
		return false, err
	}

	if referee.ReferrerID != 0 {
		return true, nil // If referral is already set, just return
	} else if referee.ID == referrer.ID || referrer.ReferrerID == referee.ID {
		return false, UserValidationError{Message: "user can not refer himself"}
	} else if ingr.VerifyURL != "" && !referee.IsVerified {
		return false, UserValidationError{Message: "referee is not verified"}
	}

	// Set reward
	refereeReward, referrerReward, err := u.setReward(ingr.RewardAmount, ingr.MaxReferral, referrer.ID)
	if err != nil {
		return false, err
	}

	// Give referral reward
	if err = u.userRepo.GiveReferralReward(GiveReferralRewardRequestIngredients{
		RefereeDeviceID:  referee.DeviceID,
		RefereeReward:    refereeReward,
		ReferrerDeviceID: referrer.DeviceID,
		ReferrerReward:   referrerReward,
		TitleForReferral: ingr.TitleForReferral,
		JWT:              ingr.JWT,
	}); err != nil {
		return false, err
	}

	// Update referee's referrer ID
	if err = u.userRepo.UpdateUserReferrerID(referee.ID, referrer.ID); err != nil {
		return false, err
	}
	return true, nil
}

// ***************************** Helpers *****************************

func (u *useCase) createUserForDevice(deviceID int64, appID int64, udt string, verifyURL string) (*DeviceUser, error) {
	if verifyURL != "" {
		if err := u.userRepo.IsVerifiedDevice(verifyURL, udt); err != nil {
			return nil, err
		}
	}
	isVerified := verifyURL != ""

	code, err := u.generateCode(strconv.FormatInt(deviceID, 10))
	if err != nil {
		return nil, err
	}

	return u.userRepo.CreateUser(deviceID, code, isVerified)
}

func (u *useCase) setReward(rewardAmount int, maxReferral int, referrerID int64) (refereeReward int, referrerReward int, err error) {
	refereeReward = rewardAmount
	referrerReward = rewardAmount
	referralCount, err := u.userRepo.GetReferralCountByUser(referrerID)
	if err != nil {
		return
	}
	if maxReferral > 0 && referralCount >= maxReferral {
		referrerReward = 0 // No more reward
	}
	return
}

func (u *useCase) generateCode(valDec string) (string, error) {
	val, err := baseconv.Convert(valDec, baseconv.DigitsDec, codeEcodingChars)
	if err != nil {
		return "", err
	}

	// Pad it with random combination of paddingChars
	if len(val) < codeMinDigit {
		rand.Seed(time.Now().UnixNano())
		var letterRunes = []rune(codePaddingChars)
		b := make([]rune, codeMinDigit-1)

		for i := range b {
			b[i] = letterRunes[rand.Intn(len(letterRunes))]
		}

		val = string(b) + val
		val = val[len(val)-codeMinDigit:]
	}

	// Return XXXX-XXXX format
	return val[:4] + "-" + val[4:], nil
}
