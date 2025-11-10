package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/jackthomas00/polaris/internal/gateway"
	"github.com/jackthomas00/polaris/internal/gateway/graphql/generated"
)

func main() {
	identityAddr := os.Getenv("IDENTITY_SVC_ADDR")
	if identityAddr == "" {
		identityAddr = "identity-svc:50051"
	}

	usageAddr := os.Getenv("USAGE_SVC_ADDR")
	if usageAddr == "" {
		usageAddr = "usage-svc:50052"
	}

	billingAddr := os.Getenv("BILLING_SVC_ADDR")
	if billingAddr == "" {
		billingAddr = "billing-svc:50053"
	}

	resolver := gateway.NewResolver(identityAddr, usageAddr, billingAddr)

	// Create GraphQL handler
	srv := handler.NewDefaultServer(
		generated.NewExecutableSchema(generated.Config{
			Resolvers: resolver,
		}),
	)

	// Add auth middleware
	authHandler := gateway.AuthMiddleware(identityAddr, srv)

	// GraphQL playground (for development)
	http.Handle("/playground", playground.Handler("GraphQL playground", "/query"))

	// GraphQL endpoint
	http.Handle("/query", authHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("api-gateway listening on :%s", port)
	log.Printf("GraphQL playground: http://localhost:%s/playground", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
