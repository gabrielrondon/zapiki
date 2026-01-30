package main

import (
	"log"
	"time"

	"github.com/gabrielrondon/zapiki/internal/api"
	"github.com/gabrielrondon/zapiki/internal/api/handlers"
	"github.com/gabrielrondon/zapiki/internal/api/middleware"
	"github.com/gabrielrondon/zapiki/internal/api/routes"
	"github.com/gabrielrondon/zapiki/internal/config"
	"github.com/gabrielrondon/zapiki/internal/prover"
	"github.com/gabrielrondon/zapiki/internal/prover/commitment"
	"github.com/gabrielrondon/zapiki/internal/service"
	"github.com/gabrielrondon/zapiki/internal/storage/postgres"
	"github.com/gabrielrondon/zapiki/internal/storage/redis"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize PostgreSQL
	pgStore, err := postgres.New(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgStore.Close()
	log.Println("Connected to PostgreSQL")

	// Initialize Redis
	redisStore, err := redis.New(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisStore.Close()
	log.Println("Connected to Redis")

	// Initialize repositories
	proofRepo := postgres.NewProofRepository(pgStore)
	apiKeyRepo := postgres.NewAPIKeyRepository(pgStore)

	// Initialize proof system factory
	factory := prover.NewFactory()

	// Register proof systems based on config
	if cfg.Proof.EnableCommitment {
		commitmentProver, err := commitment.NewCommitmentProver()
		if err != nil {
			log.Fatalf("Failed to create commitment prover: %v", err)
		}
		if err := factory.Register(commitmentProver); err != nil {
			log.Fatalf("Failed to register commitment prover: %v", err)
		}
		log.Println("Registered commitment proof system")
	}

	// TODO: Register other proof systems (Groth16, PLONK, STARK) when enabled

	// Initialize services
	proofService := service.NewProofService(factory, proofRepo)
	verifyService := service.NewVerifyService(factory)

	// Initialize handlers
	proofHandler := handlers.NewProofHandler(proofService)
	verifyHandler := handlers.NewVerifyHandler(verifyService)
	systemHandler := handlers.NewSystemHandler(factory, pgStore, redisStore)

	// Initialize middleware
	authMiddleware := middleware.NewAuth(apiKeyRepo)
	rateLimiter := redis.NewRateLimiter(redisStore)
	rateLimitMiddleware := middleware.NewRateLimit(rateLimiter, 1*time.Minute)

	// Setup router
	router := routes.NewRouter(&routes.RouterConfig{
		ProofHandler:   proofHandler,
		VerifyHandler:  verifyHandler,
		SystemHandler:  systemHandler,
		AuthMiddleware: authMiddleware,
		RateLimiter:    rateLimitMiddleware,
	})

	// Create and start server
	server := api.NewServer(&cfg.Server, router)
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
