package bfmetadata

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Embedding bfconvert.bat located in the bftools directory.
//
//go:embed bftools/bfconvert.bat
var bfconvertBat []byte

// PrintHelp executes the embedded bfconvert.bat with the --help flag and returns the output.
func PrintHelp() (string, error) {
	// Create a temporary file to hold the bfconvert.bat script
	tempDir, err := os.MkdirTemp("", "bfconvert")

	if err != nil {
		return "", fmt.Errorf("error creating temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "bfconvert.bat")

	// Write the embedded bfconvertBat to the temporary file
	err = os.WriteFile(tempFile, bfconvertBat, 0755)
	if err != nil {
		return "", fmt.Errorf("error writing bfconvert.bat to temp file: %w", err)
	}

	// Debugging: Print the path of the temporary file
	fmt.Printf("Using temporary BFCONVERTPATH: %s\n", tempFile)

	// Prepare the bfconvert.bat --help command
	cmd := exec.Command("cmd", "/C", tempFile, "--help")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Execute the command
	err = cmd.Run()
	if err != nil && !strings.Contains(out.String(), "To convert a file between formats") {
		return out.String(), fmt.Errorf("error executing bfconvert.bat --help: %w, raw stderr: %s", err, stderr.String())
	}

	// Return the help message
	return out.String(), nil
}
