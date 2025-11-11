package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

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

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Handle("/metrics", promhttp.Handler())

	// GraphQL Playground (dev only)
	if os.Getenv("ENABLE_PLAYGROUND") == "1" {
		r.Handle("/playground", playground.Handler("GraphQL playground", "/graphql"))
	}

	// Protected GraphQL endpoint
	r.Group(func(protected chi.Router) {
		protected.Use(func(next http.Handler) http.Handler {
			return gateway.AuthMiddleware(identityAddr, next)
		})
		protected.Handle("/graphql", srv)
	})

	log.Printf("api-gateway listening on %s", ":8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
