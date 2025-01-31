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

// Embed bf.bat
//
//go:embed bftools/bf.bat
var bfBat []byte

// Embed config.bat
//
//go:embed bftools/config.bat
var configBat []byte

// PrintHelp executes the bfconvert.bat with the --help flag and returns the output.
func PrintHelp() (string, error) {
	tempDir, err := prepareFiles()
	if err != nil {
		return "", err
	}

	batFile := filepath.Join(tempDir, "bfconvert.bat")

	// Set environment and execute the command using the persistent temp files
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

	return out.String(), nil
}

// GetEssentialMetadata extracts metadata from a given file using bfconvert.bat with the -nopix flag
func GetEssentialMetadata(filePath string) (string, error) {
	tempDir, err := prepareFiles()
	if err != nil {
		return "", err
	}

	batFile := filepath.Join(tempDir, "bfconvert.bat")

	// Prepare the command to execute bfconvert.bat with -nopix to extract metadata
	cmd := exec.Command("cmd", "/C", batFile, filePath, "-nopix")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Set the BF_DIR environment variable
	cmd.Env = append(os.Environ(), fmt.Sprintf("BF_DIR=%s", tempDir))

	// Execute the command
	err = cmd.Run()
	if err != nil {
		return out.String(), fmt.Errorf("error executing bfconvert.bat to get metadata: %w, stderr: %s", err, stderr.String())
	}

	return out.String(), nil
}

// prepareFiles ensures the necessary files are present in a designated temp directory.
func prepareFiles() (string, error) {
	tempDir := filepath.Join(os.TempDir(), "bioformats")
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		err = os.MkdirAll(tempDir, 0755)
		if err != nil {
			return "", fmt.Errorf("error creating temp directory: %w", err)
		}
	}

	files := map[string][]byte{
		"bfconvert.bat":          bfconvertBat,
		"bioformats_package.jar": bioformatsJar,
		"bf.bat":                 bfBat,
		"config.bat":             configBat,
	}

	for filename, data := range files {
		path := filepath.Join(tempDir, filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err := os.WriteFile(path, data, 0644)
			if err != nil {
				return "", fmt.Errorf("error writing %s to temp file: %w", filename, err)
			}
		}
	}

	return tempDir, nil
}
