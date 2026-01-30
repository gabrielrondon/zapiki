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
	"github.com/gabrielrondon/zapiki/internal/prover/snark/gnark"
	"github.com/gabrielrondon/zapiki/internal/prover/stark"
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
	jobRepo := postgres.NewJobRepository(pgStore)
	circuitRepo := postgres.NewCircuitRepository(pgStore)
	templateRepo := postgres.NewTemplateRepository(pgStore)

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

	if cfg.Proof.EnableGroth16 {
		groth16Prover := gnark.NewGroth16Prover()
		if err := factory.Register(groth16Prover); err != nil {
			log.Fatalf("Failed to register Groth16 prover: %v", err)
		}
		log.Println("Registered Groth16 proof system")
	}

	if cfg.Proof.EnablePLONK {
		plonkProver := gnark.NewPLONKProver()
		if err := factory.Register(plonkProver); err != nil {
			log.Fatalf("Failed to register PLONK prover: %v", err)
		}
		log.Println("Registered PLONK proof system")
	}

	if cfg.Proof.EnableSTARK {
		starkProver := stark.NewSTARKProver()
		if err := factory.Register(starkProver); err != nil {
			log.Fatalf("Failed to register STARK prover: %v", err)
		}
		log.Println("Registered STARK proof system")
	}

	// Initialize queue client (optional for API server)
	// The API can enqueue jobs, but the worker processes them
	// For now, pass nil and handle sync proofs only in API
	// When async is needed, the worker will process them

	// Initialize services
	proofService := service.NewProofService(factory, proofRepo, jobRepo, nil)
	verifyService := service.NewVerifyService(factory)
	circuitService := service.NewCircuitService(factory, circuitRepo)
	templateService := service.NewTemplateService(templateRepo, circuitRepo, proofService)

	// Initialize handlers
	proofHandler := handlers.NewProofHandler(proofService)
	verifyHandler := handlers.NewVerifyHandler(verifyService)
	systemHandler := handlers.NewSystemHandler(factory, pgStore, redisStore)
	jobHandler := handlers.NewJobHandler(jobRepo)
	circuitHandler := handlers.NewCircuitHandler(circuitService)
	templateHandler := handlers.NewTemplateHandler(templateService)

	// Initialize middleware
	authMiddleware := middleware.NewAuth(apiKeyRepo)
	rateLimiter := redis.NewRateLimiter(redisStore)
	rateLimitMiddleware := middleware.NewRateLimit(rateLimiter, 1*time.Minute)

	// Setup router
	router := routes.NewRouter(&routes.RouterConfig{
		ProofHandler:    proofHandler,
		VerifyHandler:   verifyHandler,
		SystemHandler:   systemHandler,
		JobHandler:      jobHandler,
		CircuitHandler:  circuitHandler,
		TemplateHandler: templateHandler,
		AuthMiddleware:  authMiddleware,
		RateLimiter:     rateLimitMiddleware,
	})

	// Create and start server
	server := api.NewServer(&cfg.Server, router)
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
