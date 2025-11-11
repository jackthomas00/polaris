package main

import (
	"log"
	"net"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	"github.com/jackthomas00/polaris/internal/billing"
	"github.com/jackthomas00/polaris/pkg/db"
	billingv1 "github.com/jackthomas00/polaris/proto/billingv1"
)

func main() {
	dsn := os.Getenv("BILLING_DB_DSN")
	if dsn == "" {
		log.Fatal("BILLING_DB_DSN not set")
	}

	pg, err := db.Connect(dsn)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}

	store := billing.NewStore(pg)
	svc := billing.NewService(store)

	// Start HTTP server for health and metrics
	go func() {
		httpMux := http.NewServeMux()
		httpMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
		httpMux.Handle("/metrics", promhttp.Handler())

		httpAddr := ":9093"
		log.Printf("billing-svc HTTP server listening on %s", httpAddr)
		if err := http.ListenAndServe(httpAddr, httpMux); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	grpcServer := grpc.NewServer()
	billingv1.RegisterBillingServer(grpcServer, svc)

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	log.Println("billing-svc listening on :50053")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
