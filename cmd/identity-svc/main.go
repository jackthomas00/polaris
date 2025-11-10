package main

import (
	"log"
	"net"
	"os"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	"github.com/jackthomas00/polaris/internal/identity"
	"github.com/jackthomas00/polaris/pkg/db"
	identityv1 "github.com/jackthomas00/polaris/proto/identityv1"
)

func main() {
	dsn := os.Getenv("IDENTITY_DB_DSN")
	if dsn == "" {
		log.Fatal("IDENTITY_DB_DSN not set")
	}

	pg, err := db.Connect(dsn)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}

	store := identity.NewStore(pg)
	svc := identity.NewService(store)

	grpcServer := grpc.NewServer()
	identityv1.RegisterIdentityServer(grpcServer, svc)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	log.Println("identity-svc listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
