package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	billingv1 "github.com/jackthomas00/polaris/proto/billingv1"
)

type Service struct {
	store *Store
	billingv1.UnimplementedBillingServer
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

func (s *Service) GenerateInvoice(ctx context.Context, req *billingv1.GenerateInvoiceRequest) (*billingv1.Invoice, error) {
	periodStart := time.Unix(req.PeriodStartUnix, 0).UTC()
	periodEnd := time.Unix(req.PeriodEndUnix, 0).UTC()

	// Get all plans for this org
	plans, err := s.store.GetPlansByOrg(ctx, req.OrgId)
	if err != nil {
		return nil, err
	}

	var totalAmount float64

	// For each plan, calculate usage and cost
	for _, plan := range plans {
		usage, err := s.store.GetUsageTotal(ctx, req.OrgId, plan.Metric, periodStart, periodEnd)
		if err != nil {
			return nil, err
		}

		// Calculate cost: max(0, usage - free_quota) * unit_price
		chargeableUsage := usage - plan.FreeQuota
		if chargeableUsage < 0 {
			chargeableUsage = 0
		}
		cost := float64(chargeableUsage) * plan.UnitPrice
		totalAmount += cost
	}

	// Create or update invoice
	invoiceID := fmt.Sprintf("inv-%s", uuid.New().String()[:8])
	invoice := &Invoice{
		ID:          invoiceID,
		OrgID:       req.OrgId,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		TotalAmount: totalAmount,
		Status:      "draft",
	}

	if err := s.store.CreateInvoice(ctx, invoice); err != nil {
		return nil, err
	}

	return &billingv1.Invoice{
		Id:              invoice.ID,
		OrgId:           invoice.OrgID,
		PeriodStartUnix: invoice.PeriodStart.Unix(),
		PeriodEndUnix:   invoice.PeriodEnd.Unix(),
		TotalAmount:     invoice.TotalAmount,
		Status:          invoice.Status,
	}, nil
}

func (s *Service) ListInvoices(ctx context.Context, req *billingv1.ListInvoicesRequest) (*billingv1.ListInvoicesResponse, error) {
	invoices, err := s.store.ListInvoices(ctx, req.OrgId)
	if err != nil {
		return nil, err
	}

	resp := &billingv1.ListInvoicesResponse{}
	for _, inv := range invoices {
		resp.Invoices = append(resp.Invoices, &billingv1.Invoice{
			Id:              inv.ID,
			OrgId:           inv.OrgID,
			PeriodStartUnix: inv.PeriodStart.Unix(),
			PeriodEndUnix:   inv.PeriodEnd.Unix(),
			TotalAmount:     inv.TotalAmount,
			Status:          inv.Status,
		})
	}
	return resp, nil
}
