package identity

import (
	"context"
	"database/sql"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

type APIKey struct {
	ID    string
	OrgID string
	Key   string
}

type Organization struct {
	ID   string
	Name string
}

func (s *Store) ValidateAPIKey(ctx context.Context, apiKey string) (*APIKey, error) {
	var ak APIKey
	err := s.db.QueryRowContext(ctx, `
		SELECT id, org_id, key
		FROM api_keys
		WHERE key = $1
	`, apiKey).Scan(&ak.ID, &ak.OrgID, &ak.Key)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ak, nil
}

func (s *Store) GetOrganization(ctx context.Context, orgID string) (*Organization, error) {
	var org Organization
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name
		FROM organizations
		WHERE id = $1
	`, orgID).Scan(&org.ID, &org.Name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &org, nil
}
