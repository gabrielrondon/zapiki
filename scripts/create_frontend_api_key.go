package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	// Get database URL from env or use default
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgresql://postgres:lFUKjRoZMsoZovfhlbmdBiCkMDsvkjZO@hopper.proxy.rlwy.net:56899/railway?sslmode=require"
	}

	ctx := context.Background()

	// Connect to database
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer conn.Close(ctx)

	// Create frontend user
	userID := "00000000-0000-0000-0000-000000000001"
	_, err = conn.Exec(ctx, `
		INSERT INTO users (id, email, name, created_at)
		VALUES ($1, 'frontend@zapiki.app', 'Frontend Application', NOW())
		ON CONFLICT (id) DO NOTHING
	`, userID)
	if err != nil {
		log.Printf("Warning creating user: %v", err)
	}

	// Check if API key already exists
	var existingKey string
	var rateLimit int
	err = conn.QueryRow(ctx,
		"SELECT key, rate_limit FROM api_keys WHERE user_id = $1 AND is_active = true LIMIT 1",
		userID).Scan(&existingKey, &rateLimit)

	if err == pgx.ErrNoRows {
		// Generate new API key
		apiKey := "zapiki_frontend_key_e49924e1831c8ea9c1be90b9b33232ad9609141ea2b180f42c8ea1dab3872933"

		// Insert new key
		_, err = conn.Exec(ctx, `
			INSERT INTO api_keys (user_id, key, name, rate_limit, is_active, created_at)
			VALUES ($1, $2, 'Frontend Application Key', 1000, true, NOW())
		`, userID, apiKey)
		if err != nil {
			log.Fatal("Failed to create API key:", err)
		}

		fmt.Println("========================================")
		fmt.Println("✅ API Key Created Successfully!")
		fmt.Println("========================================")
		fmt.Printf("\nAPI_KEY=%s\n", apiKey)
		fmt.Printf("RATE_LIMIT=%d requests/minute\n", 1000)
		fmt.Printf("USER=Frontend Application\n")
		fmt.Println("\n⚠️  IMPORTANT: Save this key securely!")
		fmt.Println("This key has high rate limits for frontend use.")
		fmt.Println("========================================")

	} else if err != nil {
		log.Fatal("Database error:", err)
	} else {
		fmt.Println("========================================")
		fmt.Println("✅ API Key Already Exists!")
		fmt.Println("========================================")
		fmt.Printf("\nAPI_KEY=%s\n", existingKey)
		fmt.Printf("RATE_LIMIT=%d requests/minute\n", rateLimit)
		fmt.Printf("USER=Frontend Application\n")
		fmt.Println("========================================")
	}
}
