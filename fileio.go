package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// WritePolarisFile writes a PolarisDocument to a .Program.polaris file with FILE_VALIDATION header
func WritePolarisFile(filename string, doc *PolarisDocument) error {
	jsonBytes, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	jsonStr := string(jsonBytes)

	hash := sha256.Sum256([]byte(jsonStr))
	hexHash := hex.EncodeToString(hash[:])

	content := fmt.Sprintf("//FILE_VALIDATION=%s\n%s\n", hexHash, jsonStr)

	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	return nil
}

// ReadPolarisFile reads and parses a .Program.polaris file
func ReadPolarisFile(filename string) (*PolarisDocument, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	content := string(data)

	// Split first line (FILE_VALIDATION) from JSON content
	idx := strings.Index(content, "\n")
	if idx < 0 {
		return nil, fmt.Errorf("invalid file format: no newline found")
	}

	firstLine := content[:idx]
	jsonContent := strings.TrimSpace(content[idx+1:])

	// Verify FILE_VALIDATION
	if strings.HasPrefix(firstLine, "//FILE_VALIDATION=") {
		expectedHash := strings.TrimPrefix(firstLine, "//FILE_VALIDATION=")
		actualHash := sha256.Sum256([]byte(jsonContent))
		actualHex := hex.EncodeToString(actualHash[:])
		if expectedHash != actualHex {
			fmt.Println("[WARNING] FILE_VALIDATION hash mismatch - file may have been modified externally")
		}
	}

	var doc PolarisDocument
	if err := json.Unmarshal([]byte(jsonContent), &doc); err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	return &doc, nil
}
