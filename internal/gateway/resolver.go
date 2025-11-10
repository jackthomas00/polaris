// This file will be overwritten by gqlgen, but we keep the resolver logic here
package gateway

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	billingv1 "github.com/jackthomas00/polaris/proto/billingv1"
	identityv1 "github.com/jackthomas00/polaris/proto/identityv1"
	usagev1 "github.com/jackthomas00/polaris/proto/usagev1"
)

type Resolver struct {
	identityAddr string
	usageAddr    string
	billingAddr  string
}

func NewResolver(identityAddr, usageAddr, billingAddr string) *Resolver {
	return &Resolver{
		identityAddr: identityAddr,
		usageAddr:    usageAddr,
		billingAddr:  billingAddr,
	}
}

func (r *Resolver) getIdentityClient() (identityv1.IdentityClient, func() error, error) {
	conn, err := grpc.NewClient(r.identityAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return identityv1.NewIdentityClient(conn), conn.Close, nil
}

func (r *Resolver) getUsageClient() (usagev1.UsageClient, func() error, error) {
	conn, err := grpc.NewClient(r.usageAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return usagev1.NewUsageClient(conn), conn.Close, nil
}

func (r *Resolver) getBillingClient() (billingv1.BillingClient, func() error, error) {
	conn, err := grpc.NewClient(r.billingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return billingv1.NewBillingClient(conn), conn.Close, nil
}

func (r *Resolver) Me(ctx context.Context) (*Organization, error) {
	authCtx := GetAuthContext(ctx)
	if authCtx == nil {
		return nil, fmt.Errorf("unauthorized")
	}

	client, close, err := r.getIdentityClient()
	if err != nil {
		return nil, err
	}
	defer close()

	resp, err := client.GetOrganization(ctx, &identityv1.GetOrganizationRequest{
		OrgId: authCtx.OrgID,
	})
	if err != nil {
		return nil, err
	}

	return &Organization{
		ID:   resp.OrgId,
		Name: resp.Name,
	}, nil
}

func (r *Resolver) Usage(ctx context.Context, metric string) ([]*UsageAggregate, error) {
	authCtx := GetAuthContext(ctx)
	if authCtx == nil {
		return nil, fmt.Errorf("unauthorized")
	}

	client, close, err := r.getUsageClient()
	if err != nil {
		return nil, err
	}
	defer close()

	resp, err := client.GetUsageSummary(ctx, &usagev1.GetUsageSummaryRequest{
		OrgId:  authCtx.OrgID,
		Metric: metric,
	})
	if err != nil {
		return nil, err
	}

	var aggregates []*UsageAggregate
	for _, agg := range resp.Aggregates {
		aggregates = append(aggregates, &UsageAggregate{
			Metric:      agg.Metric,
			Total:       float64(agg.Total),
			PeriodStart: time.Unix(agg.PeriodStartUnix, 0).UTC().Format(time.RFC3339),
			PeriodEnd:   time.Unix(agg.PeriodEndUnix, 0).UTC().Format(time.RFC3339),
		})
	}

	// If no aggregates, return empty list (not an error)
	return aggregates, nil
}

func (r *Resolver) Invoices(ctx context.Context) ([]*Invoice, error) {
	authCtx := GetAuthContext(ctx)
	if authCtx == nil {
		return nil, fmt.Errorf("unauthorized")
	}

	client, close, err := r.getBillingClient()
	if err != nil {
		return nil, err
	}
	defer close()

	resp, err := client.ListInvoices(ctx, &billingv1.ListInvoicesRequest{
		OrgId: authCtx.OrgID,
	})
	if err != nil {
		return nil, err
	}

	var invoices []*Invoice
	for _, inv := range resp.Invoices {
		invoices = append(invoices, &Invoice{
			ID:          inv.Id,
			TotalAmount: inv.TotalAmount,
			Status:      inv.Status,
			PeriodStart: time.Unix(inv.PeriodStartUnix, 0).UTC().Format(time.RFC3339),
			PeriodEnd:   time.Unix(inv.PeriodEndUnix, 0).Format(time.RFC3339),
		})
	}

	return invoices, nil
}

func (r *Resolver) RecordUsage(ctx context.Context, metric string, quantity int) (bool, error) {
	authCtx := GetAuthContext(ctx)
	if authCtx == nil {
		return false, fmt.Errorf("unauthorized")
	}

	client, close, err := r.getUsageClient()
	if err != nil {
		return false, err
	}
	defer close()

	resp, err := client.RecordUsage(ctx, &usagev1.RecordUsageRequest{
		OrgId:          authCtx.OrgID,
		Metric:         metric,
		Quantity:       int64(quantity),
		TimestampUnix:  time.Now().Unix(),
		IdempotencyKey: "", // Could generate one from request
	})
	if err != nil {
		return false, err
	}

	return resp.Success, nil
}

func (r *Resolver) GenerateInvoice(ctx context.Context, periodStart, periodEnd string) (*Invoice, error) {
	authCtx := GetAuthContext(ctx)
	if authCtx == nil {
		return nil, fmt.Errorf("unauthorized")
	}

	// Parse period start and end from RFC3339 format
	startTime, err := time.Parse(time.RFC3339, periodStart)
	if err != nil {
		return nil, fmt.Errorf("invalid periodStart format: %w", err)
	}

	endTime, err := time.Parse(time.RFC3339, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("invalid periodEnd format: %w", err)
	}

	client, close, err := r.getBillingClient()
	if err != nil {
		return nil, err
	}
	defer close()

	resp, err := client.GenerateInvoice(ctx, &billingv1.GenerateInvoiceRequest{
		OrgId:           authCtx.OrgID,
		PeriodStartUnix: startTime.Unix(),
		PeriodEndUnix:   endTime.Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &Invoice{
		ID:          resp.Id,
		TotalAmount: resp.TotalAmount,
		Status:      resp.Status,
		PeriodStart: time.Unix(resp.PeriodStartUnix, 0).UTC().Format(time.RFC3339),
		PeriodEnd:   time.Unix(resp.PeriodEndUnix, 0).UTC().Format(time.RFC3339),
	}, nil
}

type Organization struct {
	ID   string
	Name string
}

type UsageAggregate struct {
	Metric      string
	Total       float64
	PeriodStart string
	PeriodEnd   string
}

type Invoice struct {
	ID          string
	TotalAmount float64
	Status      string
	PeriodStart string
	PeriodEnd   string
}

type contextKey string

const authContextKey contextKey = "auth"

func WithAuthContext(ctx context.Context, authCtx *AuthContext) context.Context {
	return context.WithValue(ctx, authContextKey, authCtx)
}

func GetAuthContext(ctx context.Context) *AuthContext {
	authCtx, _ := ctx.Value(authContextKey).(*AuthContext)
	return authCtx
}
