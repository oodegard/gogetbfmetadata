package github.com/oodegard/gogetbfmetadata/bfmetadata

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PrintHelp executes bfconvert.bat with the --help flag and returns the output
// PrintHelp executes the bundled bfconvert.bat with the --help flag and returns the output
func PrintHelp() (string, error) {
	// Define relative path to bfconvert.bat
	executablePath := filepath.Join("bftools", "bfconvert.bat")

	// Resolve the absolute path of bfconvert.bat
	absolutePath, err := filepath.Abs(executablePath)
	if err != nil {
		return "", fmt.Errorf("error resolving bfconvert path: %w", err)
	}

	// Check if the file exists at absolutePath
	if _, err := os.Stat(absolutePath); err != nil {
		return "", fmt.Errorf("bfconvert.bat not found at path: %s", absolutePath)
	}

	// Debugging: Print the resolved path
	fmt.Printf("Using BFCONVERTPATH: %s\n", absolutePath)

	// Prepare the bfconvert.bat --help command
	cmd := exec.Command("cmd", "/C", absolutePath, "--help")
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
