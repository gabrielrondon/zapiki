// Package main demonstrates usage of the Zapiki Go SDK
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gabrielrondon/zapiki/pkg/client"
)

func main() {
	// Create client
	c := client.NewClient(
		"https://zapiki-production.up.railway.app",
		"test_zapiki_key_1230ab3c044056686e2552fb5a2648cd",
	).WithTimeout(60 * time.Second)

	ctx := context.Background()

	// Example 1: Check API health
	fmt.Println("=== Example 1: Health Check ===")
	health, err := c.Health(ctx)
	if err != nil {
		log.Fatalf("Health check failed: %v", err)
	}
	fmt.Printf("API Status: %v\n\n", health["status"])

	// Example 2: List proof systems
	fmt.Println("=== Example 2: List Proof Systems ===")
	systems, err := c.ListSystems(ctx)
	if err != nil {
		log.Fatalf("Failed to list systems: %v", err)
	}
	for _, sys := range systems {
		fmt.Printf("- %s: %v\n", sys.Name, sys.Capabilities.Features)
	}
	fmt.Println()

	// Example 3: Generate a commitment proof
	fmt.Println("=== Example 3: Generate Commitment Proof ===")
	proofReq := &client.GenerateProofRequest{
		ProofSystem: client.ProofSystemCommitment,
		Data: client.DataInput{
			Type:  "string",
			Value: "Hello from Zapiki Go SDK!",
		},
	}

	proofResp, err := c.GenerateProof(ctx, proofReq)
	if err != nil {
		log.Fatalf("Failed to generate proof: %v", err)
	}
	fmt.Printf("Proof ID: %s\n", proofResp.ProofID)
	fmt.Printf("Status: %s\n", proofResp.Status)
	fmt.Printf("Generation time: %dms\n\n", proofResp.GenerationTimeMs)

	// Example 4: Verify the proof
	fmt.Println("=== Example 4: Verify Proof ===")
	verifyReq := &client.VerifyRequest{
		ProofSystem:     client.ProofSystemCommitment,
		Proof:           proofResp.Proof,
		VerificationKey: proofResp.VerificationKey,
	}

	verifyResp, err := c.VerifyProof(ctx, verifyReq)
	if err != nil {
		log.Fatalf("Failed to verify proof: %v", err)
	}
	fmt.Printf("Valid: %v\n\n", verifyResp.Valid)

	// Example 5: List templates
	fmt.Println("=== Example 5: List Templates ===")
	templates, err := c.ListTemplates(ctx)
	if err != nil {
		log.Fatalf("Failed to list templates: %v", err)
	}
	for _, tmpl := range templates {
		if tmpl.IsActive {
			fmt.Printf("- %s (%s)\n", tmpl.Name, tmpl.Category)
		}
	}
	fmt.Println()

	// Example 6: Use a template
	if len(templates) > 0 {
		fmt.Println("=== Example 6: Generate Proof from Template ===")
		tmpl := templates[0]
		fmt.Printf("Using template: %s\n", tmpl.Name)

		// Use the example inputs from the template
		templateResp, err := c.GenerateFromTemplate(ctx, tmpl.ID, tmpl.ExampleInputs)
		if err != nil {
			log.Fatalf("Failed to generate from template: %v", err)
		}

		fmt.Printf("Proof ID: %s\n", templateResp.ProofID)
		fmt.Printf("Status: %s\n", templateResp.Status)

		if templateResp.Status == "pending" {
			fmt.Printf("Job ID: %s (async proof - poll for status)\n", templateResp.JobID)
		}
		fmt.Println()
	}

	// Example 7: Poll for async proof status
	if len(systems) > 0 {
		// Find an async system (Groth16, PLONK, or STARK)
		var asyncSystem client.ProofSystem
		for _, sys := range systems {
			if sys.Name != "commitment" {
				asyncSystem = client.ProofSystem(sys.Name)
				break
			}
		}

		if asyncSystem != "" {
			fmt.Printf("=== Example 7: Generate Async Proof (%s) ===\n", asyncSystem)
			asyncReq := &client.GenerateProofRequest{
				ProofSystem: asyncSystem,
				Data: client.DataInput{
					Type:  "json",
					Value: map[string]interface{}{"a": 5, "b": 6, "c": 30},
				},
			}

			asyncResp, err := c.GenerateProof(ctx, asyncReq)
			if err != nil {
				log.Printf("Failed to generate async proof: %v", err)
			} else {
				fmt.Printf("Proof ID: %s\n", asyncResp.ProofID)
				fmt.Printf("Status: %s\n", asyncResp.Status)

				if asyncResp.Status == "pending" {
					fmt.Printf("Job ID: %s\n", asyncResp.JobID)
					fmt.Println("Poll GET /api/v1/proofs/{proof_id} for status")
				}
			}
		}
	}

	fmt.Println("\n=== All examples completed successfully! ===")
}
