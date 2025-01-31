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

// Embed bfconvert.bat
//
//go:embed bftools/bfconvert.bat
var bfconvertBat []byte

// Embed bioformats_package.jar
//
//go:embed bftools/bioformats_package.jar
var bioformatsJar []byte

// PrintHelp executes the embedded bfconvert.bat with the --help flag and returns the output.
func PrintHelp() (string, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "bfconvert")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Write bfconvert.bat to the temporary file
	batFile := filepath.Join(tempDir, "bfconvert.bat")
	err = os.WriteFile(batFile, bfconvertBat, 0755)
	if err != nil {
		return "", fmt.Errorf("error writing bfconvert.bat to temp file: %w", err)
	}

	// Write bioformats_package.jar to the temporary file
	jarFile := filepath.Join(tempDir, "bioformats_package.jar")
	err = os.WriteFile(jarFile, bioformatsJar, 0644)
	if err != nil {
		return "", fmt.Errorf("error writing bioformats_package.jar to temp file: %w", err)
	}

	// Debugging: Print the paths
	fmt.Printf("Using temporary BFCONVERTPATH: %s\n", batFile)
	fmt.Printf("Using temporary BIOFORMATSPATH: %s\n", jarFile)

	// Adjust environment variables to point to the temp directory
	cmd := exec.Command("cmd", "/C", batFile, "--help")
	cmd.Env = append(os.Environ(), fmt.Sprintf("BF_DIR=%s", tempDir))

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
