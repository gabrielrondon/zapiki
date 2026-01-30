// Package client provides a Go SDK for the Zapiki API
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a Zapiki API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Zapiki client
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WithTimeout sets a custom timeout for the HTTP client
func (c *Client) WithTimeout(timeout time.Duration) *Client {
	c.httpClient.Timeout = timeout
	return c
}

// ProofSystem represents a proof system type
type ProofSystem string

const (
	ProofSystemCommitment ProofSystem = "commitment"
	ProofSystemGroth16    ProofSystem = "groth16"
	ProofSystemPLONK      ProofSystem = "plonk"
	ProofSystemSTARK      ProofSystem = "stark"
)

// GenerateProofRequest represents a request to generate a proof
type GenerateProofRequest struct {
	ProofSystem  ProofSystem            `json:"proof_system"`
	Data         DataInput              `json:"data"`
	PublicInputs []string               `json:"public_inputs,omitempty"`
	Options      map[string]interface{} `json:"options,omitempty"`
}

// DataInput represents input data for proof generation
type DataInput struct {
	Type  string      `json:"type"`  // "string", "json", "bytes"
	Value interface{} `json:"value"` // actual data
}

// GenerateProofResponse represents the response from proof generation
type GenerateProofResponse struct {
	ProofID          string                 `json:"proof_id"`
	Status           string                 `json:"status"`
	Proof            map[string]interface{} `json:"proof,omitempty"`
	VerificationKey  map[string]interface{} `json:"verification_key,omitempty"`
	GenerationTimeMs int64                  `json:"generation_time_ms,omitempty"`
	JobID            string                 `json:"job_id,omitempty"`
	Message          string                 `json:"message,omitempty"`
}

// VerifyRequest represents a request to verify a proof
type VerifyRequest struct {
	ProofSystem     ProofSystem            `json:"proof_system"`
	Proof           map[string]interface{} `json:"proof"`
	VerificationKey map[string]interface{} `json:"verification_key"`
	PublicInputs    []string               `json:"public_inputs,omitempty"`
}

// VerifyResponse represents the response from verification
type VerifyResponse struct {
	Valid bool   `json:"valid"`
	Error string `json:"error_message,omitempty"`
}

// SystemInfo represents information about a proof system
type SystemInfo struct {
	Name         string       `json:"name"`
	Capabilities Capabilities `json:"capabilities"`
}

// Capabilities represents proof system capabilities
type Capabilities struct {
	SupportsSetup          bool     `json:"supports_setup"`
	RequiresTrustedSetup   bool     `json:"requires_trusted_setup"`
	SupportsCustomCircuits bool     `json:"supports_custom_circuits"`
	AsyncOnly              bool     `json:"async_only"`
	TypicalGenerationTime  int64    `json:"typical_generation_time"`
	MaxProofSize           int64    `json:"max_proof_size"`
	Features               []string `json:"features"`
}

// Template represents a proof template
type Template struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Category        string                 `json:"category"`
	ProofSystem     string                 `json:"proof_system"`
	InputSchema     map[string]interface{} `json:"input_schema"`
	ExampleInputs   map[string]interface{} `json:"example_inputs"`
	Documentation   string                 `json:"documentation"`
	IsActive        bool                   `json:"is_active"`
}

// GenerateProof generates a proof
func (c *Client) GenerateProof(ctx context.Context, req *GenerateProofRequest) (*GenerateProofResponse, error) {
	resp := &GenerateProofResponse{}
	err := c.doRequest(ctx, "POST", "/api/v1/proofs", req, resp)
	return resp, err
}

// GetProof retrieves a proof by ID
func (c *Client) GetProof(ctx context.Context, proofID string) (*GenerateProofResponse, error) {
	resp := &GenerateProofResponse{}
	err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/proofs/%s", proofID), nil, resp)
	return resp, err
}

// VerifyProof verifies a proof
func (c *Client) VerifyProof(ctx context.Context, req *VerifyRequest) (*VerifyResponse, error) {
	resp := &VerifyResponse{}
	err := c.doRequest(ctx, "POST", "/api/v1/verify", req, resp)
	return resp, err
}

// ListSystems lists available proof systems
func (c *Client) ListSystems(ctx context.Context) ([]SystemInfo, error) {
	var response struct {
		Systems []SystemInfo `json:"systems"`
	}
	err := c.doRequest(ctx, "GET", "/api/v1/systems", nil, &response)
	return response.Systems, err
}

// ListTemplates lists available templates
func (c *Client) ListTemplates(ctx context.Context) ([]Template, error) {
	var response struct {
		Templates []Template `json:"templates"`
	}
	err := c.doRequest(ctx, "GET", "/api/v1/templates", nil, &response)
	return response.Templates, err
}

// GenerateFromTemplate generates a proof using a template
func (c *Client) GenerateFromTemplate(ctx context.Context, templateID string, inputs map[string]interface{}) (*GenerateProofResponse, error) {
	req := map[string]interface{}{
		"inputs": inputs,
	}
	resp := &GenerateProofResponse{}
	err := c.doRequest(ctx, "POST", fmt.Sprintf("/api/v1/templates/%s/generate", templateID), req, resp)
	return resp, err
}

// Health checks the API health
func (c *Client) Health(ctx context.Context) (map[string]interface{}, error) {
	var response map[string]interface{}
	err := c.doRequest(ctx, "GET", "/health", nil, &response)
	return response, err
}

// doRequest performs an HTTP request
func (c *Client) doRequest(ctx context.Context, method, path string, body, result interface{}) error {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	// Perform request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
