package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/mitchellh/go-homedir"
)

// Config holds the authentication tokens
type Config struct {
	BlueskySession BlueskySession `json:"bluesky_session"`
}

// BlueskySession holds Bluesky session information
type BlueskySession struct {
	AccessJwt  string `json:"access_jwt"`
	RefreshJwt string `json:"refresh_jwt"`
	Handle     string `json:"handle"`
	Did        string `json:"did"`
}

// BlueskyAuthResponse represents the response from Bluesky authentication
type BlueskyAuthResponse struct {
	AccessJwt  string `json:"accessJwt"`
	RefreshJwt string `json:"refreshJwt"`
	Did        string `json:"did"`
}

func initConfigs() error {
	if err := loadEnv(); err != nil {
		return fmt.Errorf("error loading environment variables: %w", err)
	}

	// Verify required environment variables
	required := []string{
		"BLUESKY_IDENTIFIER",
		"BLUESKY_APP_PASSWORD",
	}

	for _, env := range required {
		if os.Getenv(env) == "" {
			return fmt.Errorf("required environment variable %s is not set", env)
		}
	}

	return nil
}

func loadEnv() error {
	// Try to load from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found or error loading it: %v\n", err)
	}

	// Verify required environment variables
	required := []string{
		"BLUESKY_IDENTIFIER",
		"BLUESKY_APP_PASSWORD",
	}
	
	for _, env := range required {
		if os.Getenv(env) == "" {
			return fmt.Errorf("required environment variable %s is not set", env)
		}
	}
	return nil
}

func LoadConfig() (*Config, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "shout")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config.json")
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func SaveConfig(config *Config) error {
	home, err := homedir.Dir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "shout")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func authenticateBluesky() error {
	identifier := os.Getenv("BLUESKY_IDENTIFIER")
	appPassword := os.Getenv("BLUESKY_APP_PASSWORD")

	// Create session with Bluesky
	authURL := "https://bsky.social/xrpc/com.atproto.server.createSession"
	authReqBody, err := json.Marshal(map[string]string{
		"identifier": identifier,
		"password":   appPassword,
	})
	if err != nil {
		return fmt.Errorf("failed to encode auth request: %w", err)
	}

	authReq, err := http.NewRequest("POST", authURL, bytes.NewBuffer(authReqBody))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}
	authReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	authResp, err := client.Do(authReq)
	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}
	defer authResp.Body.Close()

	if authResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(authResp.Body)
		return fmt.Errorf("authentication failed: status %d, response: %s", authResp.StatusCode, string(bodyBytes))
	}

	var authResult BlueskyAuthResponse
	if err := json.NewDecoder(authResp.Body).Decode(&authResult); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	// Save the session
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	config.BlueskySession = BlueskySession{
		AccessJwt:  authResult.AccessJwt,
		RefreshJwt: authResult.RefreshJwt,
		Handle:     identifier,
		Did:        authResult.Did,
	}

	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Successfully authenticated with Bluesky as @%s!\n", identifier)
	return nil
}

func PostToBluesky(message string) error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if config.BlueskySession.AccessJwt == "" {
		return fmt.Errorf("not authenticated with Bluesky, please run 'auth bluesky' first")
	}

	// Create post with Bluesky
	postURL := "https://bsky.social/xrpc/com.atproto.repo.createRecord"
	postReqBody, err := json.Marshal(map[string]interface{}{
		"repo":       config.BlueskySession.Did,
		"collection": "app.bsky.feed.post",
		"record": map[string]interface{}{
			"text":      message,
			"createdAt": time.Now().Format(time.RFC3339),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to encode post request: %w", err)
	}

	postReq, err := http.NewRequest("POST", postURL, bytes.NewBuffer(postReqBody))
	if err != nil {
		return fmt.Errorf("failed to create post request: %w", err)
	}
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Authorization", "Bearer "+config.BlueskySession.AccessJwt)

	client := &http.Client{}
	postResp, err := client.Do(postReq)
	if err != nil {
		return fmt.Errorf("post request failed: %w", err)
	}
	defer postResp.Body.Close()

	if postResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(postResp.Body)
		return fmt.Errorf("post failed: status %d, response: %s", postResp.StatusCode, string(bodyBytes))
	}

	fmt.Println("Successfully posted to Bluesky!")
	return nil
}

func main() {
	if err := initConfigs(); err != nil {
		fmt.Printf("Error initializing configs: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: shout <command> [args...]")
		fmt.Println("Commands:")
		fmt.Println("  auth bluesky - Authenticate with Bluesky")
		fmt.Println("  post <message> - Post a message to Bluesky")
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "auth":
		if len(os.Args) < 3 {
			fmt.Println("Usage: shout auth <service>")
			fmt.Println("Services: bluesky")
			os.Exit(1)
		}

		service := os.Args[2]
		switch service {
		case "bluesky":
			if err := authenticateBluesky(); err != nil {
				fmt.Printf("Error authenticating with Bluesky: %v\n", err)
				os.Exit(1)
			}
		default:
			fmt.Printf("Unknown service: %s\n", service)
			fmt.Println("Supported services: bluesky")
			os.Exit(1)
		}

	case "post":
		if len(os.Args) < 3 {
			fmt.Println("Usage: shout post <message>")
			os.Exit(1)
		}

		message := os.Args[2]
		if err := PostToBluesky(message); err != nil {
			fmt.Printf("Error posting to Bluesky: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Supported commands: auth, post")
		os.Exit(1)
	}
}
