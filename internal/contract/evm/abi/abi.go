package abi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rxtech-lab/smart-contract-cli/internal/errors"
)

type AbiArray []ABIElement

type AbiObject struct {
	Abi      AbiArray
	Bytecode string
	Metadata map[string]any
}

// ParseAbi parse an abi string which can be in array or object format or
// Abi object format and returns an AbiArray.
func ParseAbi(abi string) (AbiArray, error) {
	var abiArray AbiArray
	var abiObject AbiObject

	err := json.Unmarshal([]byte(abi), &abiArray)
	// check if error is not nil, try to unmarshal as an object
	if err != nil {
		err = json.Unmarshal([]byte(abi), &abiObject)
		if err != nil {
			return nil, errors.WrapABIError(err, errors.ErrCodeInvalidABIFormat, "failed to parse ABI: invalid JSON format")
		}
		return abiObject.Abi, nil
	}

	return abiArray, nil
}

// ReadAbi reads an ABI from a local file path or download it from a remote source.
func ReadAbi(filepath string) (AbiArray, error) {
	// Check if filepath is a URL
	if strings.HasPrefix(filepath, "http://") || strings.HasPrefix(filepath, "https://") {
		return downloadAbi(filepath)
	}

	// Otherwise, treat it as a local file path
	return readAbiFromFile(filepath)
}

// downloadAbi downloads an ABI from a remote URL and parses it.
func downloadAbi(url string) (AbiArray, error) {
	// Create HTTP client
	client := &http.Client{}

	// Create request with context
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, errors.WrapABIError(err, errors.ErrCodeABIParseFailed, fmt.Sprintf("failed to create request for URL: %s", url))
	}

	// Set headers
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WrapABIError(err, errors.ErrCodeABIParseFailed, fmt.Sprintf("failed to download ABI from URL: %s", url))
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log error but don't fail the operation
			_ = closeErr
		}
	}()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewABIError(errors.ErrCodeABIParseFailed, fmt.Sprintf("failed to download ABI: received status code %d from URL: %s", resp.StatusCode, url))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WrapABIError(err, errors.ErrCodeABIParseFailed, fmt.Sprintf("failed to read response body from URL: %s", url))
	}

	// Parse ABI
	return ParseAbi(string(body))
}

// readAbiFromFile reads an ABI from a local file and parses it.
func readAbiFromFile(filePath string) (AbiArray, error) {
	// Validate filePath to prevent directory traversal
	cleaned := filepath.Clean(filePath)
	if strings.Contains(cleaned, "..") {
		return nil, errors.NewABIError(errors.ErrCodeABIParseFailed, fmt.Sprintf("invalid file path: %s", filePath))
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.WrapABIError(err, errors.ErrCodeABIParseFailed, fmt.Sprintf("failed to read ABI file: %s", filePath))
	}

	// Parse ABI
	return ParseAbi(string(data))
}
