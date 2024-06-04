package repo

import (
	"context"

	authsvc "github.com/Buzzvil/buzzapis/go/auth"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/auth"
)

// New creates auth repository.
func New(client authsvc.AuthServiceClient) *Repository {
	return &Repository{client: client}
}

// Repository struct
type Repository struct {
	client authsvc.AuthServiceClient
}

// CreateAuth calls CreateAuth to authsvc
func (r Repository) CreateAuth(identifier auth.Identifier) (string, error) {
	req := authsvc.CreateAuthRequest{
		Identifier: &authsvc.Identifier{
			AppId:           identifier.AppID,
			PublisherUserId: identifier.PublisherUserID,
			Ifa:             identifier.IFA,
		},
	}

	res, err := r.client.CreateAuth(context.Background(), &req)
	if err != nil {
		return "", err
	}

	return res.GetToken(), nil
}

// GetAuth calls GetAuth to authsvc
func (r Repository) GetAuth(token string) (*auth.Auth, error) {
	req := authsvc.GetAuthRequest{Token: token}
	at, err := r.client.GetAuth(context.Background(), &req)
	if err != nil {
		return nil, err
	}

	a := &auth.Auth{
		AccountID:       at.AccountId,
		AppID:           at.Identifier.AppId,
		PublisherUserID: at.Identifier.PublisherUserId,
		IFA:             at.Identifier.Ifa,
	}

	return a, nil
}
