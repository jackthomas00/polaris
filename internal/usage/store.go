package usage

import (
	"context"
	"database/sql"
	"time"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) InsertUsageEvent(ctx context.Context, orgID, metric string, quantity int64, ts time.Time, idemKey string) error {
	// Only enforce uniqueness if idempotency_key is provided
	if idemKey != "" {
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO usage_events (org_id, metric, quantity, occurred_at, idempotency_key)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (org_id, idempotency_key) DO NOTHING
		`, orgID, metric, quantity, ts, idemKey)
		return err
	}
	// No idempotency key, just insert
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO usage_events (org_id, metric, quantity, occurred_at, idempotency_key)
		VALUES ($1, $2, $3, $4, NULL)
	`, orgID, metric, quantity, ts)
	return err
}

type Aggregate struct {
	Metric      string
	Total       int64
	PeriodStart time.Time
	PeriodEnd   time.Time
}

func (s *Store) GetAggregates(ctx context.Context, orgID, metric string) ([]Aggregate, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT metric, total, period_start, period_end
		FROM usage_aggregates
		WHERE org_id = $1 AND metric = $2
		ORDER BY period_start DESC
		LIMIT 30
	`, orgID, metric)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Aggregate
	for rows.Next() {
		var a Aggregate
		if err := rows.Scan(&a.Metric, &a.Total, &a.PeriodStart, &a.PeriodEnd); err != nil {
			return nil, err
		}
		res = append(res, a)
	}
	return res, rows.Err()
}
