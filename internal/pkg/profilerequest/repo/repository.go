package repo

import (
	"context"
	"time"

	pb "github.com/Buzzvil/buzzapis/go/profile"
	pr "github.com/Buzzvil/buzzscreen-api/internal/pkg/profilerequest"
)

// Repository struct definition
type Repository struct {
	grpcClient pb.ProfileServiceClient
}

// PopulateProfile populates profile UID table based on ID types received inside pixel log data.
func (r *Repository) PopulateProfile(account pr.Account) error {
	ctx, cancel := context.WithTimeout(context.Background(), millisecondFactor*time.Millisecond)
	defer cancel()

	req := r.buildProfileIDRequest(account)
	if !r.isProfileValid(req.Profile) {
		return nil
	}

	_, err := r.grpcClient.GetProfileID(ctx, req)
	if err != nil {
		return err
	}
	return err
}

func (r *Repository) buildProfileIDRequest(account pr.Account) *pb.GetProfileIDRequest {
	return &pb.GetProfileIDRequest{Profile: &pb.Profile{
		AppId:     account.AppID,
		Ifa:       account.IFA,
		AccountId: account.AccountID,
		CookieId:  account.CookieID,
		AppUserId: account.AppUserID}}
}

func (r *Repository) isProfileValid(profile *pb.Profile) bool {
	return profile.AppId != 0 ||
		profile.AppUserId != "" ||
		profile.Ifa != "" ||
		profile.AccountId != 0 ||
		profile.CookieId != "" ||
		profile.UserId != "" ||
		profile.Email != "" ||
		profile.Fingerprint != ""
}

// New returns new profilerequest repository
func New(grpcClient pb.ProfileServiceClient) *Repository {
	return &Repository{
		grpcClient: grpcClient,
	}
}
