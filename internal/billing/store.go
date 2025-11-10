package billing

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

type Plan struct {
	ID        string
	OrgID     string
	Name      string
	Metric    string
	UnitPrice float64
	FreeQuota int64
}

type Invoice struct {
	ID          string
	OrgID       string
	PeriodStart time.Time
	PeriodEnd   time.Time
	TotalAmount float64
	Status      string
}

func (s *Store) GetPlansByOrg(ctx context.Context, orgID string) ([]Plan, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, org_id, name, metric, unit_price, free_quota
		FROM plans
		WHERE org_id = $1
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []Plan
	for rows.Next() {
		var p Plan
		if err := rows.Scan(&p.ID, &p.OrgID, &p.Name, &p.Metric, &p.UnitPrice, &p.FreeQuota); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	return plans, rows.Err()
}

func (s *Store) GetUsageTotal(ctx context.Context, orgID, metric string, start, end time.Time) (int64, error) {
	var total int64
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(quantity), 0)
		FROM usage_events
		WHERE org_id = $1 AND metric = $2 
		AND occurred_at >= $3 AND occurred_at < $4
	`, orgID, metric, start, end).Scan(&total)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return total, nil
}

func (s *Store) CreateInvoice(ctx context.Context, invoice *Invoice) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO invoices (id, org_id, period_start, period_end, total_amount, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			total_amount = EXCLUDED.total_amount,
			status = EXCLUDED.status
	`, invoice.ID, invoice.OrgID, invoice.PeriodStart, invoice.PeriodEnd, invoice.TotalAmount, invoice.Status)
	return err
}

func (s *Store) ListInvoices(ctx context.Context, orgID string) ([]Invoice, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, org_id, period_start, period_end, total_amount, status
		FROM invoices
		WHERE org_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []Invoice
	for rows.Next() {
		var inv Invoice
		if err := rows.Scan(&inv.ID, &inv.OrgID, &inv.PeriodStart, &inv.PeriodEnd, &inv.TotalAmount, &inv.Status); err != nil {
			return nil, err
		}
		invoices = append(invoices, inv)
	}
	return invoices, rows.Err()
}
