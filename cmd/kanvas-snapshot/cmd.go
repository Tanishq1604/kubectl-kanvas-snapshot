package kanvas_snapshot

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/layer5io/meshkit/logger"
	"github.com/meshery/kubectl-kanvas-snapshot/pkg/snapshot/config"
	"github.com/meshery/kubectl-kanvas-snapshot/pkg/snapshot/errors"
	"github.com/meshery/kubectl-kanvas-snapshot/pkg/snapshot/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	// Constants for default values
	defaultMesheryURL = "https://playground.meshery.io"
	apiEndpoint       = "/api/pattern/import" // Changed to match mesheryctl's endpoint
)

var (
	// Global variables for configuration
	ProviderToken          string
	MesheryAPIBaseURL      string
	MesheryCloudAPIBaseURL string
	WorkflowAccessToken    string
	Log                    log.Logger
	// Configuration
	Config *config.Config
)

var (
	// Command flags
	manifestPath string
	email        string
	designName   string
	recursive    bool
	skipWorkflow bool
	// GitHub workflow configuration
	repoOwner  string
	repoName   string
	branchName string
	workflowID string
)

// Regular expression for email validation
var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)

// generateKanvasSnapshotCmd represents the root command for kubectl kanvas-snapshot
var generateKanvasSnapshotCmd = &cobra.Command{
	Use:   "kanvas-snapshot",
	Short: "Generate a Kanvas snapshot using Kubernetes manifests",
	Long: `Generate a Kanvas snapshot by providing Kubernetes manifest files.

		This command allows you to generate a snapshot in Meshery using Kubernetes manifests.

		Example usage:

		kubectl kanvas-snapshot -f ./manifests/deployment.yaml -e your-email@example.com --name my-deployment
		kubectl kanvas-snapshot -f ./manifests/ --recursive --name my-project

		Flags:
		-f, --file      string	Path to Kubernetes manifest file or directory (required)
		-r, --recursive		Recursively process all manifest files in the directory
		-e, --email     string	Email address to notify when snapshot is ready (optional)
		    --name      string	(optional) Name for the Meshery design
		-h			Help for kubectl Kanvas Snapshot plugin`,

	RunE: kanvasSnapshotRunE,
}

// getManifestContents reads the manifest file(s) and returns their contents
func getManifestContents(path string, recursive bool) ([]string, error) {
	var manifests []string

	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, errors.ErrReadingManifestFile(err)
	}

	if fileInfo.IsDir() {
		manifests, err = processDirectory(path, recursive)
		if err != nil {
			return nil, errors.ErrReadingManifestFile(err)
		}
		if len(manifests) == 0 {
			return nil, errors.ErrReadingManifestFile(fmt.Errorf("no YAML files found in the specified directory"))
		}
	} else {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, errors.ErrReadingManifestFile(err)
		}
		manifests = append(manifests, string(content))
	}

	return manifests, nil
}

// processDirectory finds all YAML and YML files in a directory
func processDirectory(dirPath string, recursive bool) ([]string, error) {
	var manifests []string
	walkFn := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories if not recursive
		if info.IsDir() {
			if path != dirPath && !recursive {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file is a YAML file
		if strings.HasSuffix(strings.ToLower(path), ".yaml") || strings.HasSuffix(strings.ToLower(path), ".yml") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			manifests = append(manifests, string(content))
			Log.Infof("Added manifest file: %s", path)
		}
		return nil
	}

	err := filepath.Walk(dirPath, walkFn)
	return manifests, err
}

// MesheryDesignPayload represents the payload for creating a design in Meshery
type MesheryDesignPayload struct {
	Name       string `json:"name"`
	File       string `json:"file"` // This will hold base64 encoded content
	FileName   string `json:"file_name"`
	Email      string `json:"email,omitempty"`
	SourceType string `json:"source_type"` // Added to specify Kubernetes manifest
}

// ExtractNameFromPath extracts the name from the file path
func ExtractNameFromPath(path string) string {
	filename := filepath.Base(path)
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

// CreateMesheryDesign creates a new design in Meshery
func CreateMesheryDesign(manifest, name, email string) (string, error) {
	// Extract filename from manifestPath for the file_name field
	fileName := filepath.Base(manifestPath)

	// Base64 encode the manifest content
	encodedManifest := base64.StdEncoding.EncodeToString([]byte(manifest))

	payload := MesheryDesignPayload{
		Name:       name,
		File:       encodedManifest,
		FileName:   fileName,
		SourceType: "Kubernetes Manifest",
	}

	if email != "" {
		payload.Email = email
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		Log.Errorf("Failed to marshal payload: %v", err)
		return "", errors.ErrDecodingAPI(err)
	}

	endpoint := apiEndpoint
	if Config != nil && Config.Meshery.SnapshotEndpoint != "" {
		endpoint = Config.Meshery.SnapshotEndpoint
	}

	// Simple URL construction
	fullURL := fmt.Sprintf("%s%s", MesheryAPIBaseURL, endpoint)
	Log.Infof("Sending request to: %s", fullURL)

	// Create the request
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		Log.Errorf("Failed to create new request: %v", err)
		return "", errors.ErrHTTPPostRequest(err)
	}

	// Set content type header
	req.Header.Set("Content-Type", "application/json")

	// Add authentication if token is available
	if ProviderToken != "" {
		Log.Info("Using Meshery token for authentication")
		cookieValue := fmt.Sprintf("token=%s;meshery-provider=Meshery", ProviderToken)
		req.Header.Set("Cookie", cookieValue)
	} else {
		Log.Warn("No Meshery token provided, authentication will likely fail")
	}

	// Create a client with the cookie jar
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: time.Second * 30,
		Jar:     jar,
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		Log.Errorf("HTTP request failed: %v", err)
		return "", errors.ErrHTTPPostRequest(err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		Log.Errorf("Failed to read response body: %v", err)
		return "", errors.ErrHTTPPostRequest(err)
	}

	// Log response details
	Log.Infof("Response status: %s", resp.Status)
	Log.Debugf("Response body: %s", string(body))

	// Check if we got HTML instead of JSON (auth redirect)
	if strings.Contains(string(body), "<!DOCTYPE html>") || strings.Contains(string(body), "<html") {
		Log.Warn("Received HTML response instead of JSON - authentication failed")
		return "", errors.ErrHTTPPostRequest(fmt.Errorf("authentication failed"))
	}

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		Log.Warnf("Unexpected response code: %d", resp.StatusCode)
		return "", errors.ErrHTTPPostRequest(fmt.Errorf("unexpected response code: %d", resp.StatusCode))
	}

	// First, try to parse the response as a JSON object
	var designID string
	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		// If that fails, try to parse it as a JSON array
		var responseArray []interface{}
		if err := json.Unmarshal(body, &responseArray); err != nil {
			Log.Errorf("Failed to parse response JSON: %v", err)
			return "", errors.ErrDecodingAPI(err)
		}

		// If it's an array and has at least one element
		if len(responseArray) > 0 {
			// Try to get the ID from the first element if it's an object
			if firstItem, ok := responseArray[0].(map[string]interface{}); ok {
				if id, ok := firstItem["id"].(string); ok {
					designID = id
					Log.Infof("Successfully created Meshery design. ID: %s", designID)
				}
			}
		}
	} else {
		// Extract the design ID from the object response
		if id, ok := responseData["id"].(string); ok {
			designID = id
			Log.Infof("Successfully created Meshery design. ID: %s", designID)
		} else if id, ok := responseData["pattern_file"].(map[string]interface{})["id"].(string); ok {
			// Try another common response format
			designID = id
			Log.Infof("Successfully created Meshery design. ID: %s", designID)
		} else if id, ok := responseData["pattern_id"].(string); ok {
			// Try another field name
			designID = id
			Log.Infof("Successfully created Meshery design. ID: %s", designID)
		}
	}

	// If we couldn't extract an ID but the request was successful
	if designID == "" {
		Log.Warnf("Could not extract design ID from response: %s", trimString(string(body), 200))
		return "", errors.ErrDecodingAPI(fmt.Errorf("could not extract design ID from response"))
	}

	return designID, nil
}

// Helper function to trim strings for logging
func trimString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// GenerateSnapshot publishes the design to Meshery's pattern catalog
func GenerateSnapshot(designID, assetLocation, token string) error {
	if token == "" {
		Log.Warn("GITHUB_TOKEN environment variable not set. Snapshot generation will be skipped.")
		Log.Info("Please set GITHUB_TOKEN environment variable to trigger GitHub workflow.")
		return nil
	}

	// Generate direct URL to view in Meshery
	mesheryViewURL := getDesignViewURL(designID)
	Log.Infof("View your design in Meshery: %s", mesheryViewURL)

	// Set default values for GitHub repository and workflow
	repoOwnerValue := repoOwner
	if repoOwnerValue == "" {
		repoOwnerValue = "layer5labs"
		Log.Infof("No repository owner specified, using default: %s", repoOwnerValue)
	}

	repoNameValue := repoName
	if repoNameValue == "" {
		repoNameValue = "kubectl-kanvas-snapshot"
		Log.Infof("No repository name specified, using default: %s", repoNameValue)
	}

	workflowIDValue := workflowID
	if workflowIDValue == "" {
		workflowIDValue = "kanvas.yaml"
		Log.Infof("No workflow ID specified, using default: %s", workflowIDValue)
	}

	// If assetLocation is not provided, generate a default one
	if assetLocation == "" {
		assetLocation = fmt.Sprintf("https://raw.githubusercontent.com/layer5labs/meshery-extensions-packages/master/action-assets/kubectl-plugin-assets/%s.png", designID)
		Log.Infof("Using default asset location: %s", assetLocation)
	}

	// Trigger GitHub workflow using REST API
	Log.Info("Triggering GitHub workflow to generate snapshot...")

	// Construct the GitHub API URL to trigger workflow
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/workflows/%s/dispatches",
		repoOwnerValue, repoNameValue, workflowIDValue)

	// Prepare payload for workflow dispatch
	payload := map[string]interface{}{
		"ref": "master", // or any branch where the workflow is defined
		"inputs": map[string]string{
			"designID":      designID, // Changed from contentID to designID
			"assetLocation": assetLocation,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		Log.Errorf("Failed to marshal payload: %v", err)
		return errors.ErrGeneratingSnapshot(err)
	}

	// Create the request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		Log.Errorf("Failed to create request: %v", err)
		return errors.ErrGeneratingSnapshot(err)
	}

	// Set headers for GitHub API
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Set("Content-Type", "application/json")

	// Create HTTP client
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		Log.Errorf("Failed to trigger workflow: %v", err)
		return errors.ErrGeneratingSnapshot(err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		Log.Errorf("Workflow trigger failed with status %d: %s", resp.StatusCode, string(body))
		return errors.ErrGeneratingSnapshot(fmt.Errorf("workflow trigger failed with status %d: %s", resp.StatusCode, string(body)))
	}

	Log.Info("Workflow triggered successfully!")
	Log.Infof("Your design snapshot will be available at: %s", assetLocation)
	Log.Info("This process may take a few minutes to complete...")

	return nil
}

// isValidEmail validates an email address format
func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// Main is the entrypoint for the plugin
func Main(providerToken, mesheryCloudAPIBaseURL, mesheryAPIBaseURL, workflowAccessToken string) {
	// Set global variables
	ProviderToken = providerToken
	MesheryCloudAPIBaseURL = mesheryCloudAPIBaseURL
	MesheryAPIBaseURL = mesheryAPIBaseURL
	WorkflowAccessToken = workflowAccessToken

	// Initialize logger
	setupLogger()

	// Load configuration
	var err error
	Config, err = config.LoadConfig()
	if err != nil {
		Log.Errorf("Failed to load configuration: %v", err)
	} else {
		Log.Infof("Loaded configuration from: %s", config.GetConfigFilePath())
		// Override API URL from config if set
		if Config.Meshery.URL != "" {
			Log.Infof("Using Meshery URL from config: %s", Config.Meshery.URL)
			MesheryAPIBaseURL = Config.Meshery.URL
		}
	}

	// Setup command flags
	generateKanvasSnapshotCmd.Flags().StringVarP(&manifestPath, "file", "f", "", "Path to the Kubernetes manifest file (required)")
	generateKanvasSnapshotCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Process manifest files recursively in directories")
	generateKanvasSnapshotCmd.Flags().StringVarP(&designName, "name", "n", "", "Name for the Meshery design (default: extracted from manifest path)")
	generateKanvasSnapshotCmd.Flags().StringVarP(&email, "email", "e", "", "Email address for notifications")
	generateKanvasSnapshotCmd.Flags().BoolVarP(&skipWorkflow, "skip-workflow", "s", false, "Skip publishing to Meshery's pattern catalog")
	generateKanvasSnapshotCmd.Flags().StringVarP(&MesheryAPIBaseURL, "meshery-url", "m", "", "Meshery API URL (default: http://localhost:9081)")
	generateKanvasSnapshotCmd.Flags().StringVarP(&ProviderToken, "meshery-token", "t", "", "Meshery authentication token")

	// GitHub workflow configuration flags
	generateKanvasSnapshotCmd.Flags().StringVar(&repoOwner, "repo-owner", "", "GitHub repository owner (defaults to layer5labs)")
	generateKanvasSnapshotCmd.Flags().StringVar(&repoName, "repo-name", "", "GitHub repository name (defaults to meshery-extensions-packages)")
	generateKanvasSnapshotCmd.Flags().StringVar(&branchName, "branch", "", "GitHub repository branch (defaults to master)")
	generateKanvasSnapshotCmd.Flags().StringVar(&workflowID, "workflow", "", "GitHub workflow ID (defaults to kanvas.yaml)")

	// Mark required flags
	_ = generateKanvasSnapshotCmd.MarkFlagRequired("file")

	// Update flag descriptions
	generateKanvasSnapshotCmd.Flags().SetAnnotation("file", "required", []string{"true"})
	generateKanvasSnapshotCmd.Flags().SetAnnotation("name", "help", []string{"Name for the Meshery design. If not provided, will be extracted from the manifest path."})
	generateKanvasSnapshotCmd.Flags().SetAnnotation("email", "help", []string{"Email address for notifications when the design is ready."})
	generateKanvasSnapshotCmd.Flags().SetAnnotation("recursive", "help", []string{"Process manifest files recursively in directories."})
	generateKanvasSnapshotCmd.Flags().SetAnnotation("skip-workflow", "help", []string{"Skip publishing to Meshery's pattern catalog. The design will still be created but won't be published."})
	generateKanvasSnapshotCmd.Flags().SetAnnotation("meshery-url", "help", []string{"Meshery API URL. Defaults to http://localhost:9081 if not set."})
	generateKanvasSnapshotCmd.Flags().SetAnnotation("meshery-token", "help", []string{"Meshery authentication token. Can also be set via MESHERY_TOKEN environment variable."})

	// Execute the command
	if err := generateKanvasSnapshotCmd.Execute(); err != nil {
		Log.Error(fmt.Errorf("%v", err))
		os.Exit(1)
	}
}

// setupLogger initializes the logger
func setupLogger() {
	// Initialize logger with meshkit
	mesheryLogger, err := logger.New("kubectl-kanvas-snapshot", logger.Options{
		Format:   logger.TerminalLogFormat,
		LogLevel: int(logrus.DebugLevel),
	})

	if err != nil {
		// Fall back to simple logger if meshkit logger initialization fails
		Log = log.SetupLogger("kubectl-kanvas-snapshot", true, os.Stdout)
		Log.Warn(fmt.Sprintf("Failed to initialize meshkit logger: %v. Using fallback logger.", err))
	} else {
		Log = &log.MeshkitLogger{Log: mesheryLogger}
	}
}

// init runs before Main and initializes the command
func init() {
	// This empty init function is needed to satisfy Go's initialization requirements
}

// Update how we generate URLs for viewing designs
func getDesignViewURL(designID string) string {
	// Meshery UI URLs are structured as /extension/meshmap?mode=design&design=<designID>
	return fmt.Sprintf("%s/extension/meshmap?mode=design&design=%s",
		strings.TrimSuffix(MesheryAPIBaseURL, "/api"), designID)
}

// RunE function for the command
func kanvasSnapshotRunE(_ *cobra.Command, _ []string) error {
	// Check if Meshery token is set
	if ProviderToken == "" {
		Log.Warn("MESHERY_TOKEN environment variable not set. Working in offline mode.")
		Log.Info("Please set the MESHERY_TOKEN environment variable to use online features.")
		Log.Info("You can obtain a token from your Meshery or Meshery Cloud profile.")
	}

	// Check if Meshery API URL is set
	if MesheryAPIBaseURL == "" {
		Log.Warn("Meshery API URL not set. Using default: http://localhost:9081")
		MesheryAPIBaseURL = defaultMesheryURL
	}

	// Log the endpoints being used
	endpoint := apiEndpoint
	if Config != nil && Config.Meshery.SnapshotEndpoint != "" {
		endpoint = Config.Meshery.SnapshotEndpoint
	}
	Log.Infof("Using Meshery API URL: %s", MesheryAPIBaseURL)
	Log.Infof("Using API endpoint: %s", endpoint)

	// Use the extracted name from manifest path if not provided
	if designName == "" {
		designName = ExtractNameFromPath(manifestPath)
		Log.Warnf("No design name provided. Using extracted name: %s", designName)
	}

	// Validate email if provided
	if email != "" && !isValidEmail(email) {
		return errors.ErrInvalidEmailFormat(email)
	}

	// Process manifest files
	Log.Info("Processing manifest files...")
	manifests, err := getManifestContents(manifestPath, recursive)
	if err != nil {
		return err
	}
	Log.Infof("Processed %d manifest file(s)", len(manifests))

	// Combine all manifests, ensuring proper spacing
	combinedManifest := strings.Join(manifests, "\n---\n")

	// Log manifest size for debugging
	Log.Debugf("Manifest size: %d bytes", len(combinedManifest))

	// Create Meshery Design
	Log.Info("Creating Meshery design...")
	designID, err := CreateMesheryDesign(combinedManifest, designName, email)
	if err != nil {
		Log.Errorf("Failed to create Meshery design: %v", err)
		return errors.ErrCreatingMesheryDesign(err)
	}

	// Generate direct URL to view in Meshery
	mesheryViewURL := getDesignViewURL(designID)
	Log.Infof("View your design in Meshery: %s", mesheryViewURL)

	if skipWorkflow {
		Log.Info("Skipping publishing as --skip-workflow flag is set.")
		Log.Infof("\nDesign created successfully with ID: %s", designID)
		return nil
	}

	Log.Info("Triggering GitHub workflow to generate snapshot...")
	err = GenerateSnapshot(designID, "", WorkflowAccessToken)
	if err != nil {
		return errors.ErrGeneratingSnapshot(err)
	}

	// Output success message with clear instructions
	Log.Infof("\nDesign created successfully with ID: %s", designID)
	Log.Info("GitHub workflow has been triggered to generate a snapshot.")

	// Help user understand what to do next
	if repoOwner == "" {
		repoOwner = "layer5labs"
	}
	if repoName == "" {
		repoName = "meshery"
	}
	if workflowID == "" {
		workflowID = "kanvas.yaml"
	}

	Log.Infof("To access the snapshot images:")
	Log.Infof("1. Go to https://github.com/%s/%s/actions/workflows/%s", repoOwner, repoName, workflowID)
	Log.Infof("2. Find the most recent workflow run for designID: %s", designID)
	Log.Infof("3. Wait for the workflow run to complete (~1-2 minutes)")
	Log.Infof("4. Download the 'design-screenshots' artifact from the completed workflow")

	return nil
}
