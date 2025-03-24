package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/layer5io/meshkit/logger"
	kanvas_snapshot "github.com/meshery/kubectl-kanvas-snapshot/cmd/kanvas-snapshot"
	"github.com/meshery/kubectl-kanvas-snapshot/pkg/snapshot/config"
	"github.com/meshery/kubectl-kanvas-snapshot/pkg/snapshot/log"
	"github.com/sirupsen/logrus"
)

const (
	// Environment variable names
	envMesheryToken    = "MESHERY_TOKEN"
	envMesheryCloudURL = "MESHERY_CLOUD_URL"
	envGitHubToken     = "GITHUB_TOKEN"
)

var (
	// Token for authenticating with Meshery
	providerToken string
	// URL for Meshery Cloud API
	mesheryCloudAPIBaseURL string
	// URL for Meshery API
	mesheryAPIBaseURL string
	// GitHub Personal Access Token for triggering workflow
	workflowAccessToken string
)

func main() {
	// Create logger
	mesheryLogger, err := logger.New("kubectl-kanvas-snapshot", logger.Options{
		Format:   logger.TerminalLogFormat,
		LogLevel: int(logrus.InfoLevel),
		Output:   os.Stdout,
	})

	var Log log.Logger
	if err != nil {
		// Fall back to simple logger if meshkit logger initialization fails
		Log = log.SetupLogger("kubectl-kanvas-snapshot", false, os.Stdout)
		Log.Warn(fmt.Sprintf("Failed to initialize meshkit logger: %v. Using fallback logger.", err))
	} else {
		Log = &log.MeshkitLogger{Log: mesheryLogger}
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		Log.Errorf("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	// Set API URLs from configuration
	mesheryAPIBaseURL = cfg.Meshery.URL

	// Set environment variables
	setEnvironmentVariables()

	// Try to load from .env file if variables are not set
	loadEnvFile(Log)

	// Log configuration information
	Log.Infof("Kubectl Kanvas Snapshot Plugin")
	Log.Infof("--------------------------------")
	Log.Debugf("Meshery API URL: %s", mesheryAPIBaseURL)
	Log.Debugf("Meshery Cloud API URL: %s", mesheryCloudAPIBaseURL)

	if providerToken == "" {
		Log.Warn("MESHERY_TOKEN environment variable not set. Working in offline mode.")
		Log.Warn("Please set the MESHERY_TOKEN environment variable to use online features.")
		Log.Warn("You can obtain a token from your Meshery or Meshery Cloud profile.")
	}

	if workflowAccessToken == "" {
		Log.Warn("GITHUB_TOKEN environment variable not set. Snapshot generation will be skipped.")
	}

	// Start the command handler
	kanvas_snapshot.Main(providerToken, mesheryCloudAPIBaseURL, mesheryAPIBaseURL, workflowAccessToken)
}

// loadEnvFile attempts to load environment variables from .env file
func loadEnvFile(Log log.Logger) {
	// Try to open .env file
	file, err := os.Open(".env")
	if err != nil {
		Log.Debugf("Could not open .env file: %v", err)
		return
	}
	defer file.Close()

	// Scan the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}

		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, "\"'")

		// Set variables based on key
		switch key {
		case envMesheryToken:
			if providerToken == "" {
				providerToken = value
				Log.Infof("Loaded MESHERY_TOKEN from .env file")
			}
		case envMesheryCloudURL:
			if mesheryCloudAPIBaseURL == "" {
				mesheryCloudAPIBaseURL = value
				Log.Infof("Loaded MESHERY_CLOUD_URL from .env file")
			}
		case envGitHubToken:
			if workflowAccessToken == "" {
				workflowAccessToken = value
				Log.Infof("Loaded GITHUB_TOKEN from .env file")
			}
		case "MESHERY_API_URL":
			if mesheryAPIBaseURL == "" {
				mesheryAPIBaseURL = value
				Log.Infof("Loaded MESHERY_API_URL from .env file")
			}
		}
	}

	if err := scanner.Err(); err != nil {
		Log.Warnf("Error reading .env file: %v", err)
	}
}

// setEnvironmentVariables loads variables from environment
func setEnvironmentVariables() {
	// Get provider token from environment
	providerToken = os.Getenv(envMesheryToken)

	// Get Meshery Cloud URL from environment
	mesheryCloudAPIBaseURL = os.Getenv(envMesheryCloudURL)

	// Get GitHub token from environment
	workflowAccessToken = os.Getenv(envGitHubToken)
}
