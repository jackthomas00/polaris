package graphql

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UsageAggregate struct {
	Metric      string  `json:"metric"`
	Total       float64 `json:"total"`
	PeriodStart string  `json:"periodStart"`
	PeriodEnd   string  `json:"periodEnd"`
}

type Invoice struct {
	ID          string  `json:"id"`
	TotalAmount float64 `json:"totalAmount"`
	Status      string  `json:"status"`
	PeriodStart string  `json:"periodStart"`
	PeriodEnd   string  `json:"periodEnd"`
}
