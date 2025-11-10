package usage

import (
	"context"
	"time"

	usagev1 "github.com/jackthomas00/polaris/proto/usagev1"
)

type Service struct {
	store *Store
	usagev1.UnimplementedUsageServer
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

func (s *Service) RecordUsage(ctx context.Context, req *usagev1.RecordUsageRequest) (*usagev1.RecordUsageResponse, error) {
	if req.OrgId == "" || req.Metric == "" || req.Quantity <= 0 {
		return &usagev1.RecordUsageResponse{Success: false}, nil
	}

	ts := time.Unix(req.TimestampUnix, 0).UTC()
	if ts.IsZero() {
		ts = time.Now().UTC()
	}

	if err := s.store.InsertUsageEvent(ctx, req.OrgId, req.Metric, req.Quantity, ts, req.IdempotencyKey); err != nil {
		return nil, err
	}

	return &usagev1.RecordUsageResponse{Success: true}, nil
}

func (s *Service) GetUsageSummary(ctx context.Context, req *usagev1.GetUsageSummaryRequest) (*usagev1.GetUsageSummaryResponse, error) {
	aggs, err := s.store.GetAggregates(ctx, req.OrgId, req.Metric)
	if err != nil {
		return nil, err
	}

	resp := &usagev1.GetUsageSummaryResponse{}
	for _, a := range aggs {
		resp.Aggregates = append(resp.Aggregates, &usagev1.UsageAggregate{
			Metric:          a.Metric,
			Total:           a.Total,
			PeriodStartUnix: a.PeriodStart.Unix(),
			PeriodEndUnix:   a.PeriodEnd.Unix(),
		})
	}
	return resp, nil
}
