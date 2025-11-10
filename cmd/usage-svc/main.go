package main

import (
	"log"
	"net"
	"os"

	_ "github.com/lib/pq"
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
