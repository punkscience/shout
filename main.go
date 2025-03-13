package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mitchellh/go-homedir"
)

// Maximum character count allowed for Bluesky posts
const BlueskeyCharacterLimit = 300

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

func promptForCredentials() (string, string, error) {
	var identifier, password string

	fmt.Print("Enter your Bluesky identifier (email or handle): ")
	if _, err := fmt.Scanln(&identifier); err != nil {
		return "", "", fmt.Errorf("failed to read identifier: %w", err)
	}

	fmt.Print("Enter your Bluesky app password: ")
	if _, err := fmt.Scanln(&password); err != nil {
		return "", "", fmt.Errorf("failed to read password: %w", err)
	}

	// Clean input by trimming spaces
	identifier = strings.TrimSpace(identifier)
	password = strings.TrimSpace(password)

	return identifier, password, nil
}

func authenticateWithCredentials(identifier, appPassword string) (*BlueskyAuthResponse, error) {
	// Create session with Bluesky
	authURL := "https://bsky.social/xrpc/com.atproto.server.createSession"
	authReqBody, err := json.Marshal(map[string]string{
		"identifier": identifier,
		"password":   appPassword,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode auth request: %w", err)
	}

	authReq, err := http.NewRequest("POST", authURL, bytes.NewBuffer(authReqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create auth request: %w", err)
	}
	authReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	authResp, err := client.Do(authReq)
	if err != nil {
		return nil, fmt.Errorf("authentication request failed: %w", err)
	}
	defer authResp.Body.Close()

	if authResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(authResp.Body)
		return nil, fmt.Errorf("authentication failed: status %d, response: %s", authResp.StatusCode, string(bodyBytes))
	}

	var authResult BlueskyAuthResponse
	if err := json.NewDecoder(authResp.Body).Decode(&authResult); err != nil {
		return nil, fmt.Errorf("failed to decode auth response: %w", err)
	}

	return &authResult, nil
}

func refreshBlueskyToken(refreshJwt string) (*BlueskyAuthResponse, error) {
	refreshURL := "https://bsky.social/xrpc/com.atproto.server.refreshSession"
	refreshReq, err := http.NewRequest("POST", refreshURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}
	refreshReq.Header.Set("Authorization", "Bearer "+refreshJwt)

	client := &http.Client{}
	refreshResp, err := client.Do(refreshReq)
	if err != nil {
		return nil, fmt.Errorf("refresh request failed: %w", err)
	}
	defer refreshResp.Body.Close()

	if refreshResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(refreshResp.Body)
		return nil, fmt.Errorf("token refresh failed: status %d, response: %s", refreshResp.StatusCode, string(bodyBytes))
	}

	var refreshResult BlueskyAuthResponse
	if err := json.NewDecoder(refreshResp.Body).Decode(&refreshResult); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}

	return &refreshResult, nil
}

func authenticateBluesky() error {
	// First check if we have stored tokens
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// If we have a refresh token, try to use it first
	if config.BlueskySession.RefreshJwt != "" {
		fmt.Println("Attempting to refresh existing session...")
		authResult, err := refreshBlueskyToken(config.BlueskySession.RefreshJwt)
		if err == nil {
			// Successfully refreshed tokens
			config.BlueskySession.AccessJwt = authResult.AccessJwt
			config.BlueskySession.RefreshJwt = authResult.RefreshJwt
			
			if err := SaveConfig(config); err != nil {
				return fmt.Errorf("failed to save refreshed tokens: %w", err)
			}
			
			fmt.Printf("Successfully refreshed session for @%s!\n", config.BlueskySession.Handle)
			return nil
		}
	}
	
	fmt.Println("Will try with credentials instead.")

	// Always prompt for credentials
	fmt.Println("Please enter your Bluesky credentials:")
	identifier, appPassword, err := promptForCredentials()
	if err != nil {
		return fmt.Errorf("error prompting for credentials: %w", err)
	}

	// Authenticate with provided credentials
	authResult, err := authenticateWithCredentials(identifier, appPassword)
	if err != nil {
		return err
	}

	// Save the session
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
	// Check message length against the character limit using Unicode character count
	messageLength := utf8.RuneCountInString(message)
	fmt.Printf("Your message contains %d characters (limit: %d)\n", messageLength, BlueskeyCharacterLimit)
	
	if messageLength > BlueskeyCharacterLimit {
		remainingCount := messageLength - BlueskeyCharacterLimit
		return fmt.Errorf("message exceeds Bluesky's %d character limit by %d characters. Your message has %d characters. Please shorten your message", BlueskeyCharacterLimit, remainingCount, messageLength)
	}

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

	// Check if the token is expired (status 401)
	if postResp.StatusCode == http.StatusUnauthorized {
		fmt.Println("Access token expired. Attempting to refresh...")
		
		// Try to refresh the token
		if config.BlueskySession.RefreshJwt != "" {
			authResult, err := refreshBlueskyToken(config.BlueskySession.RefreshJwt)
			if err != nil {
				return fmt.Errorf("failed to refresh token: %w, please re-authenticate with 'auth bluesky'", err)
			}
			
			// Update the tokens in config
			config.BlueskySession.AccessJwt = authResult.AccessJwt
			config.BlueskySession.RefreshJwt = authResult.RefreshJwt
			
			// Save the updated tokens
			if err := SaveConfig(config); err != nil {
				return fmt.Errorf("failed to save refreshed tokens: %w", err)
			}
			
			// Try posting again with the new token
			return PostToBluesky(message)
		}
		
		return fmt.Errorf("token expired and no refresh token available, please re-authenticate with 'auth bluesky'")
	}

	if postResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(postResp.Body)
		return fmt.Errorf("posting failed: status %d, response: %s", postResp.StatusCode, string(bodyBytes))
	}
	
	fmt.Println("Successfully posted to Bluesky!")
	return nil
}

func main() {
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
