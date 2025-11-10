package main

import (
	"log"
	"net"
	"os"

	_ "github.com/lib/pq"
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
