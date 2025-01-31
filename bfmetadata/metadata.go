func PrintHelp() (string, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "bfconvert")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Write bf.bat to the temporary file
	bfBatFile := filepath.Join(tempDir, "bf.bat")
	err = os.WriteFile(bfBatFile, bfBat, 0755)
	if err != nil {
		return "", fmt.Errorf("error writing bf.bat to temp file: %w", err)
	}

	// Write other files like bfconvert.bat and bioformats_package.jar
	batFile := filepath.Join(tempDir, "bfconvert.bat")
	err = os.WriteFile(batFile, bfconvertBat, 0755)
	if err != nil {
		return "", fmt.Errorf("error writing bfconvert.bat to temp file: %w", err)
	}

	jarFile := filepath.Join(tempDir, "bioformats_package.jar")
	err = os.WriteFile(jarFile, bioformatsJar, 0644)
	if err != nil {
		return "", fmt.Errorf("error writing bioformats_package.jar to temp file: %w", err)
	}

	// Set environment and execute the command using the temporary files
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
