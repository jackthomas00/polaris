package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	"github.com/jackthomas00/polaris/internal/usage"
	"github.com/jackthomas00/polaris/pkg/db"
	usagev1 "github.com/jackthomas00/polaris/proto/usagev1"
)

func main() {
	dsn := os.Getenv("USAGE_DB_DSN")
	if dsn == "" {
		log.Fatal("USAGE_DB_DSN not set")
	}

	pg, err := db.Connect(dsn)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}

	store := usage.NewStore(pg)
	svc := usage.NewService(store)

	// Start HTTP server for health and metrics
	go func() {
		httpMux := http.NewServeMux()
		httpMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
		httpMux.Handle("/metrics", promhttp.Handler())

		httpAddr := ":9092"
		log.Printf("usage-svc HTTP server listening on %s", httpAddr)
		if err := http.ListenAndServe(httpAddr, httpMux); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Start periodic aggregation goroutine
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		// Run immediately on startup
		ctx := context.Background()
		if err := store.AggregateUsageEvents(ctx); err != nil {
			log.Printf("aggregation error: %v", err)
		}

		// Then run every 60 seconds
		for range ticker.C {
			ctx := context.Background()
			if err := store.AggregateUsageEvents(ctx); err != nil {
				log.Printf("aggregation error: %v", err)
			}
		}
	}()

	grpcServer := grpc.NewServer()
	usagev1.RegisterUsageServer(grpcServer, svc)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	log.Println("usage-svc listening on :50052")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
