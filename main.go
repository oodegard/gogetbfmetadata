package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// Struct to hold series information
type SeriesInfo struct {
	Timepoints int
	Channels   int
	ZStacks    int
	Thumbnail  bool
}

// Struct for YAML output
type AcquisitionMetadata struct {
	Camera          string                       `yaml:"Camera"`
	Channels        map[string]map[string]string `yaml:"Channels"`
	DateAndTime     string                       `yaml:"DateAndTime"`
	ImageDimensions struct {
		C int `yaml:"C"`
		T int `yaml:"T"`
		X int `yaml:"X"`
		Y int `yaml:"Y"`
		Z int `yaml:"Z"`
	} `yaml:"Image dimensions"`
	PixelSize struct {
		X string `yaml:"X"`
		Y string `yaml:"Y"`
		Z string `yaml:"Z"`
	} `yaml:"Pixel size"`
}

type SampleInfo struct {
	Channels      map[string]map[string]string `yaml:"Channels"`
	ElabJournalID string                       `yaml:"Elab journal ID"`
}

// Function to check if Java is installed
func checkJavaInstallation() error {
	_, err := exec.LookPath("java")
	if err != nil {
		return fmt.Errorf("java is not installed or not in your path: %v", err)
	}
	return nil
}

// Function to read metadata from an image using Bio-Formats
func readImageMetadata(imagePath, jarPath string) ([]SeriesInfo, map[string]string, string, error) {
	// The "-nopix" flag tells Bio-Formats to read only the metadata, skipping pixel data
	cmd := exec.Command("java", "-cp", jarPath, "loci.formats.tools.ImageInfo", "-nopix", imagePath)

	// Run the command and capture the output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, nil, "", fmt.Errorf("error running Java command: %v", err)
	}

	// Convert output to string for parsing
	outputStr := string(output)
	fmt.Printf("outputStr: %v\n", outputStr)

	// Use regular expressions to extract relevant metadata
	seriesRegex := regexp.MustCompile(`Series\s*count\s*=\s*(\d+)`)
	sizeTRegex := regexp.MustCompile(`SizeT\s*=\s*(\d+)`)
	sizeCRegex := regexp.MustCompile(`SizeC\s*=\s*(\d+)`)
	sizeZRegex := regexp.MustCompile(`SizeZ\s*=\s*(\d+)`)
	laserRegex := regexp.MustCompile(`ExcitationWavelength\s*=\s*(\d+\s*nm)`)
	thumbnailRegex := regexp.MustCompile(`Thumbnail\s*=\s*(\S+)`) // Added regex for Thumbnail field

	// Extract values
	seriesMatches := seriesRegex.FindStringSubmatch(outputStr)
	sizeTMatches := sizeTRegex.FindAllStringSubmatch(outputStr, -1)
	sizeCMatches := sizeCRegex.FindAllStringSubmatch(outputStr, -1)
	sizeZMatches := sizeZRegex.FindAllStringSubmatch(outputStr, -1)
	laserMatches := laserRegex.FindAllStringSubmatch(outputStr, -1)
	thumbnailMatches := thumbnailRegex.FindAllStringSubmatch(outputStr, -1)

	// Parse the extracted values
	seriesCount := 0
	if len(seriesMatches) > 1 {
		seriesCount, _ = strconv.Atoi(seriesMatches[1])
	}

	// Store metadata for all series
	var seriesInfos []SeriesInfo
	for i := 0; i < seriesCount; i++ {
		sizeT := 0
		sizeC := 0
		sizeZ := 0
		thumbnail := true

		if i < len(sizeTMatches) && len(sizeTMatches[i]) > 1 {
			sizeT, _ = strconv.Atoi(sizeTMatches[i][1])
		}
		if i < len(sizeCMatches) && len(sizeCMatches[i]) > 1 {
			sizeC, _ = strconv.Atoi(sizeCMatches[i][1])
		}
		if i < len(sizeZMatches) && len(sizeZMatches[i]) > 1 {
			sizeZ, _ = strconv.Atoi(sizeZMatches[i][1])
		}
		if i < len(thumbnailMatches) && len(thumbnailMatches[i]) > 1 {
			thumbnail = (thumbnailMatches[i][1] == "true")
		}

		// Only append series where Thumbnail = true
		if thumbnail {
			seriesInfos = append(seriesInfos, SeriesInfo{
				Timepoints: sizeT,
				Channels:   sizeC,
				ZStacks:    sizeZ,
				Thumbnail:  thumbnail,
			})
		}
	}

	// Extract laser wavelength information
	laserWavelengths := map[string]string{}
	for i, match := range laserMatches {
		channelKey := fmt.Sprintf("Channel %d", i+1)
		laserWavelengths[channelKey] = match[1]
	}

	// Extract DateAndTime
	dateTimeRegex := regexp.MustCompile(`DateAndTime\s*=\s*(\S+\s*\S+)`)
	dateTimeMatch := dateTimeRegex.FindStringSubmatch(outputStr)
	dateAndTime := ""
	if len(dateTimeMatch) > 1 {
		dateAndTime = dateTimeMatch[1]
	}

	return seriesInfos, laserWavelengths, dateAndTime, nil
}

// Function to group and print series with the same metadata
func saveYAML(imagePath string, seriesInfos []SeriesInfo, laserWavelengths map[string]string, dateAndTime string) error {
	// Prepare the YAML data
	acquisitionMetadata := AcquisitionMetadata{
		Camera:      "Please_fill_in_a_camera",
		Channels:    map[string]map[string]string{},
		DateAndTime: dateAndTime,
	}

	// Fill in the channels information and image dimensions
	for i, info := range seriesInfos {
		channelKey := fmt.Sprintf("Channel %d", i+1)
		acquisitionMetadata.Channels[channelKey] = map[string]string{
			"Laser wavelength": laserWavelengths[channelKey],
		}

		acquisitionMetadata.ImageDimensions = struct {
			C int `yaml:"C"`
			T int `yaml:"T"`
			X int `yaml:"X"`
			Y int `yaml:"Y"`
			Z int `yaml:"Z"`
		}{
			C: info.Channels,
			T: info.Timepoints,
			X: 2048, // Placeholder values
			Y: 2048, // Placeholder values
			Z: info.ZStacks,
		}
		acquisitionMetadata.PixelSize = struct {
			X string `yaml:"X"`
			Y string `yaml:"Y"`
			Z string `yaml:"Z"`
		}{
			X: "0.0614",
			Y: "0.0614",
			Z: "null",
		}
	}

	// Prepare sample info
	sampleInfo := SampleInfo{
		Channels:      map[string]map[string]string{},
		ElabJournalID: "Please_fill_in_a_journal_ID",
	}
	for i := range seriesInfos {
		channelKey := fmt.Sprintf("Channel %d", i+1)
		sampleInfo.Channels[channelKey] = map[string]string{
			"Fluorophore": "Please_fill_in_a_fluorophore",
			"Sample name": "Please_fill_in_a_sample_name",
		}
	}

	// Combine into a single map
	yamlData := map[string]interface{}{
		"Acquisition metadata": acquisitionMetadata,
		"Sample info":          sampleInfo,
	}

	// Convert the YAML data to bytes
	yamlBytes, err := yaml.Marshal(yamlData)
	if err != nil {
		return fmt.Errorf("error marshalling YAML: %v", err)
	}

	// Write the YAML data to a file
	yamlFilePath := strings.TrimSuffix(imagePath, filepath.Ext(imagePath)) + ".yaml"
	err = os.WriteFile(yamlFilePath, yamlBytes, 0644)
	if err != nil {
		return fmt.Errorf("error writing YAML file: %v", err)
	}

	fmt.Printf("YAML file saved as: %s\n", yamlFilePath)
	return nil
}

// Main function
func main() {
	// Define command-line flags
	imagePath := flag.String("image", "C:/Users/Øyvind/OneDrive - Universitetet i Oslo/Work/03_UiO/04_Microscope_images_DO_NOT_USE/20240215_123_antibodyTest/20240215_123_DFCP1-GFP_IBIDI1B_ELYS_7.ims", "Path to the image file")
	jarPath := flag.String("jar", "./bioformats_package.jar", "Path to the Bioformats JAR file")
	showHelp := flag.Bool("h", false, "Show help message")

	// Parse the flags
	flag.Parse()

	// Show help message if requested
	if *showHelp {
		flag.Usage()
		return
	}

	// Check if Java is installed
	if err := checkJavaInstallation(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Read the metadata from the image file
	seriesInfos, laserWavelengths, dateAndTime, err := readImageMetadata(*imagePath, *jarPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Save the YAML file
	err = saveYAML(*imagePath, seriesInfos, laserWavelengths, dateAndTime)
	if err != nil {
		fmt.Printf("Error saving YAML: %v\n", err)
		os.Exit(1)
	}
}
