package identity

import (
	"context"

	identityv1 "github.com/jackthomas00/polaris/proto/identityv1"
)

type Service struct {
	store *Store
	identityv1.UnimplementedIdentityServer
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

func (s *Service) ValidateApiKey(ctx context.Context, req *identityv1.ValidateApiKeyRequest) (*identityv1.ValidateApiKeyResponse, error) {
	ak, err := s.store.ValidateAPIKey(ctx, req.ApiKey)
	if err != nil {
		return nil, err
	}
	if ak == nil {
		return &identityv1.ValidateApiKeyResponse{Valid: false}, nil
	}
	return &identityv1.ValidateApiKeyResponse{
		Valid: true,
		OrgId: ak.OrgID,
	}, nil
}

func (s *Service) GetOrganization(ctx context.Context, req *identityv1.GetOrganizationRequest) (*identityv1.GetOrganizationResponse, error) {
	org, err := s.store.GetOrganization(ctx, req.OrgId)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return &identityv1.GetOrganizationResponse{}, nil
	}
	return &identityv1.GetOrganizationResponse{
		OrgId: org.ID,
		Name:  org.Name,
	}, nil
}
