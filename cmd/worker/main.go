package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gabrielrondon/zapiki/internal/config"
	"github.com/gabrielrondon/zapiki/internal/prover"
	"github.com/gabrielrondon/zapiki/internal/prover/commitment"
	"github.com/gabrielrondon/zapiki/internal/prover/snark/gnark"
	"github.com/gabrielrondon/zapiki/internal/queue"
	"github.com/gabrielrondon/zapiki/internal/storage/postgres"
	"github.com/gabrielrondon/zapiki/internal/worker"
	"github.com/hibiken/asynq"
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

	// Initialize repositories
	proofRepo := postgres.NewProofRepository(pgStore)
	jobRepo := postgres.NewJobRepository(pgStore)

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

	// TODO: Register STARK when enabled

	// Initialize worker processor
	processor := worker.NewProcessor(factory, proofRepo, jobRepo)

	// Create asynq mux and register handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TypeProofGeneration, processor.HandleProofGeneration)

	// Create and start queue server
	redisAddr := cfg.Redis.Addr()
	concurrency := 10 // Number of concurrent workers

	queueServer := queue.NewServer(redisAddr, concurrency)

	log.Printf("Starting worker with %d concurrent processors", concurrency)
	log.Printf("Connected to Redis at %s", redisAddr)

	// Handle graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		log.Println("Shutting down worker...")
		queueServer.Shutdown()
		log.Println("Worker stopped")
	}()

	// Start processing
	if err := queueServer.Start(mux); err != nil {
		log.Fatalf("Worker error: %v", err)
	}
}
