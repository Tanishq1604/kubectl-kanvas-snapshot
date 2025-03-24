package errors

import (
	"fmt"

	"github.com/layer5io/meshkit/errors"
)

var (
	// ErrDecodingAPICode represents API decoding failures
	ErrDecodingAPICode = "kubectl-kanvas-snapshot-1001"
	// ErrHTTPPostRequestCode represents HTTP request failures
	ErrHTTPPostRequestCode = "kubectl-kanvas-snapshot-1002"
	// ErrUnexpectedResponseCodeCode represents unexpected HTTP response codes
	ErrUnexpectedResponseCodeCode = "kubectl-kanvas-snapshot-1003"
	// ErrCreatingMesheryDesignCode represents Meshery design creation failures
	ErrCreatingMesheryDesignCode = "kubectl-kanvas-snapshot-1004"
	// ErrInvalidEmailFormatCode represents invalid email format
	ErrInvalidEmailFormatCode = "kubectl-kanvas-snapshot-1005"
	// ErrGeneratingSnapshotCode represents snapshot generation failures
	ErrGeneratingSnapshotCode = "kubectl-kanvas-snapshot-1006"
	// ErrReadingManifestFileCode represents manifest file reading failures
	ErrReadingManifestFileCode = "kubectl-kanvas-snapshot-1007"
)

// ErrDecodingAPI returns error for API decoding failures
func ErrDecodingAPI(err error) error {
	return errors.New(ErrDecodingAPICode, errors.Alert, []string{
		fmt.Sprintf("error decoding API response: %v", err),
	}, []string{
		"Invalid or unexpected response format from Meshery API",
	}, []string{
		"Ensure Meshery API server is running the correct version",
		"Check if the response format has changed in the Meshery API",
	}, []string{})
}

// ErrHTTPPostRequest returns error for HTTP request failures
func ErrHTTPPostRequest(err error) error {
	return errors.New(ErrHTTPPostRequestCode, errors.Alert, []string{
		fmt.Sprintf("error making HTTP POST request: %v", err),
	}, []string{
		"Failed to connect to Meshery API server",
		"Network connectivity issues",
	}, []string{
		"Ensure Meshery API server is running and accessible",
		"Check network connectivity to the Meshery server",
		"Verify if the Meshery server URL is correct in the configuration",
	}, []string{})
}

// ErrUnexpectedResponseCode returns error for unexpected HTTP response codes
func ErrUnexpectedResponseCode(code int, body string) error {
	return errors.New(ErrUnexpectedResponseCodeCode, errors.Alert, []string{
		fmt.Sprintf("unexpected response code: %d, body: %s", code, body),
	}, []string{
		"Meshery API server returned an error response",
	}, []string{
		"Check if the Meshery API server is functioning correctly",
		"Verify if the request payload is valid",
		"Check if your authentication token is valid",
	}, []string{})
}

// ErrCreatingMesheryDesign returns error for Meshery design creation failures
func ErrCreatingMesheryDesign(err error) error {
	return errors.New(ErrCreatingMesheryDesignCode, errors.Alert, []string{
		fmt.Sprintf("error creating Meshery design: %v", err),
	}, []string{
		"Failed to create a new design in Meshery",
	}, []string{
		"Verify the manifest file is valid Kubernetes YAML",
		"Check if you have permissions to create designs in Meshery",
		"Ensure Meshery server is running the latest version",
	}, []string{})
}

// ErrInvalidEmailFormat returns error for invalid email format
func ErrInvalidEmailFormat(email string) error {
	return errors.New(ErrInvalidEmailFormatCode, errors.Alert, []string{
		fmt.Sprintf("invalid email format for '%s'", email),
	}, []string{
		"The provided email address format is not valid",
	}, []string{
		"Provide a valid email address in the format user@example.com",
	}, []string{})
}

// ErrGeneratingSnapshot returns error for snapshot generation failures
func ErrGeneratingSnapshot(err error) error {
	return errors.New(ErrGeneratingSnapshotCode, errors.Alert, []string{
		fmt.Sprintf("error generating snapshot: %v", err),
	}, []string{
		"Failed to trigger snapshot generation workflow",
	}, []string{
		"Check if GitHub access token is provided and valid",
		"Verify network connectivity to GitHub API",
	}, []string{})
}

// ErrReadingManifestFile returns error for manifest file reading failures
func ErrReadingManifestFile(err error) error {
	return errors.New(ErrReadingManifestFileCode, errors.Alert, []string{
		fmt.Sprintf("error reading manifest file: %v", err),
	}, []string{
		"Failed to read the specified Kubernetes manifest file",
	}, []string{
		"Ensure the file exists and has correct permissions",
		"Verify the path to the manifest file is correct",
	}, []string{})
}
