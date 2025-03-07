package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/term"
)

// Config holds the user's Bluesky credentials
type Config struct {
	Username string `json:"username"`
	Password string `json:"password"`
}


// LoadConfig loads the configuration from the config file
func LoadConfig() (*Config, error) {
	// Get the home directory
	home, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create the config path
	configDir := filepath.Join(home, ".config", "shout")
	configPath := filepath.Join(configDir, "config.json")

	// Check if the config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil // Return nil to indicate config doesn't exist
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the config
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to the config file
func SaveConfig(config *Config) error {
	// Get the home directory
	home, err := homedir.Dir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create the config directory if it doesn't exist
	configDir := filepath.Join(home, ".config", "shout")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create the config path
	configPath := filepath.Join(configDir, "config.json")

	// Marshal the config
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write the config file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// PromptCredentials prompts the user for their Bluesky credentials
func PromptCredentials() (*Config, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your Bluesky username (without the '@' prefix): ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read username: %w", err)
	}
	username = strings.TrimSpace(username)
	// Remove '@' prefix if the user included it
	username = strings.TrimPrefix(username, "@")
	fmt.Print("Enter your Bluesky password/app password (input will be hidden): ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println() // Add a newline after the password input

	return &Config{
		Username: username,
		Password: string(passwordBytes),
	}, nil
}

// AuthResponse represents the response from the Bluesky authentication endpoint
type AuthResponse struct {
	AccessJwt  string `json:"accessJwt"`
	RefreshJwt string `json:"refreshJwt"`
	Handle     string `json:"handle"`
	Did        string `json:"did"`
}

// PostToBluesky posts a message to Bluesky using direct HTTP requests
func PostToBluesky(message, username, password string) error {
	// Step 1: Authenticate with Bluesky to get a session token
	authURL := "https://bsky.social/xrpc/com.atproto.server.createSession"
	
	// Remove '@' prefix from username if present
	cleanUsername := strings.TrimPrefix(username, "@")
	
	authReqBody, err := json.Marshal(map[string]string{
		"identifier": cleanUsername,
		"password":   password,
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
	
	var authResult AuthResponse
	if err := json.NewDecoder(authResp.Body).Decode(&authResult); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}
	
	// Step 2: Create a post using the authentication token
	postURL := "https://bsky.social/xrpc/com.atproto.repo.createRecord"
	
	// Current time in RFC3339 format
	currentTime := time.Now().UTC().Format(time.RFC3339)
	
	// Prepare the post record
	postRecord := map[string]interface{}{
		"$type": "app.bsky.feed.post",
		"text":  message,
		"createdAt": currentTime,
	}
	
	postReqBody, err := json.Marshal(map[string]interface{}{
		"repo":       authResult.Did,
		"collection": "app.bsky.feed.post",
		"record":     postRecord,
	})
	if err != nil {
		return fmt.Errorf("failed to encode post request: %w", err)
	}
	
	postReq, err := http.NewRequest("POST", postURL, bytes.NewBuffer(postReqBody))
	if err != nil {
		return fmt.Errorf("failed to create post request: %w", err)
	}
	
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Authorization", "Bearer "+authResult.AccessJwt)
	
	postResp, err := client.Do(postReq)
	if err != nil {
		return fmt.Errorf("post request failed: %w", err)
	}
	defer postResp.Body.Close()
	
	if postResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(postResp.Body)
		return fmt.Errorf("post creation failed: status %d, response: %s", postResp.StatusCode, string(bodyBytes))
	}
	
	return nil
}

func main() {
	// Check if we have a command line argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: shout \"your message here to post to Bluesky\"")
		os.Exit(1)
	}

	// Get the message from the command line
	message := os.Args[1]


	// Try to load the config
	config, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// If we don't have a config, prompt for credentials
	if config == nil {
		fmt.Println("No configuration found. Please enter your Bluesky credentials.")
		config, err = PromptCredentials()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error prompting for credentials: %v\n", err)
			os.Exit(1)
		}

		// Save the config for next time
		if err := SaveConfig(config); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			// Continue anyway since we have the credentials
		} else {
			fmt.Println("Configuration saved successfully.")
		}
	}

	// Post the message to Bluesky
	displayUsername := config.Username
	if !strings.HasPrefix(displayUsername, "@") {
		displayUsername = "@" + displayUsername
	}
	fmt.Println("Posting to Bluesky as", displayUsername, "...")
	if err := PostToBluesky(message, config.Username, config.Password); err != nil {
		fmt.Fprintf(os.Stderr, "Error posting to Bluesky: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully posted to Bluesky:")
	fmt.Println(message)
}

