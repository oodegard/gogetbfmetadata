package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

// Struct to hold metadata information
type Metadata struct {
	SeriesInfos []SeriesInfo
	//LaserWavelengths map[string]string
	DateAndTime string
}

// Struct to hold series information
type SeriesInfo struct {
	C int
	T int
	X int
	Y int
	Z int
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
// Function to read metadata from an image using Bio-Formats
func readImageMetadata(imagePath, jarPath string) (*Metadata, error) {
	// The "-nopix" flag tells Bio-Formats to read only the metadata, skipping pixel data
	cmd := exec.Command("java", "-cp", jarPath, "loci.formats.tools.ImageInfo", "-nopix", imagePath)
	fmt.Printf("cmd: %v\n", cmd)
	// Run the command and capture the output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error running Java command: %v", err)
	}

	// Convert output to string for parsing
	outputStr := string(output)
	fmt.Printf("outputStr: %v\n", outputStr)

	// Use regular expressions to extract relevant metadata
	seriesRegex := regexp.MustCompile(`Series\s*count\s*=\s*(\d+)`)
	sizeTRegex := regexp.MustCompile(`SizeT\s*=\s*(\d+)`)
	sizeCRegex := regexp.MustCompile(`SizeC\s*=\s*(\d+)`)
	sizeZRegex := regexp.MustCompile(`SizeZ\s*=\s*(\d+)`)

	sizeXRegex := regexp.MustCompile(`Width\s*=\s*(\d+)`)
	sizeYRegex := regexp.MustCompile(`Height\s*=\s*(\d+)`)

	//laserRegex := regexp.MustCompile(`ExcitationWavelength\s*=\s*(\d+\s*nm)`)
	thumbnailRegex := regexp.MustCompile(`Thumbnail\s*series\s*=\s*(true|false)`)

	// Extract values
	seriesMatches := seriesRegex.FindStringSubmatch(outputStr)
	sizeTMatches := sizeTRegex.FindAllStringSubmatch(outputStr, -1)
	sizeCMatches := sizeCRegex.FindAllStringSubmatch(outputStr, -1)
	sizeZMatches := sizeZRegex.FindAllStringSubmatch(outputStr, -1)
	//laserMatches := laserRegex.FindAllStringSubmatch(outputStr, -1)

	sizeXMatches := sizeXRegex.FindAllStringSubmatch(outputStr, -1)
	sizeYMatches := sizeYRegex.FindAllStringSubmatch(outputStr, -1)

	thumbnailMatches := thumbnailRegex.FindAllStringSubmatch(outputStr, -1)

	// Handle series count
	seriesCount := 0
	if len(seriesMatches) > 1 {
		seriesCount, _ = strconv.Atoi(seriesMatches[1])
	}

	// Store metadata for non-thumbnail series
	var seriesInfos []SeriesInfo
	for i := 0; i < seriesCount; i++ {
		if i < len(thumbnailMatches) && len(thumbnailMatches[i]) > 1 && thumbnailMatches[i][1] == "false" {
			// Only process non-thumbnail series

			sizeT, sizeC, sizeZ, sizeX, sizeY := 0, 0, 0, 0, 0

			if i < len(sizeTMatches) && len(sizeTMatches[i]) > 1 {
				sizeT, _ = strconv.Atoi(sizeTMatches[i][1])
			}
			if i < len(sizeCMatches) && len(sizeCMatches[i]) > 1 {
				sizeC, _ = strconv.Atoi(sizeCMatches[i][1])
			}
			if i < len(sizeZMatches) && len(sizeZMatches[i]) > 1 {
				sizeZ, _ = strconv.Atoi(sizeZMatches[i][1])
			}

			if i < len(sizeXMatches) && len(sizeXMatches[i]) > 1 {
				sizeX, _ = strconv.Atoi(sizeXMatches[i][1])
			}

			if i < len(sizeYMatches) && len(sizeYMatches[i]) > 1 {
				sizeY, _ = strconv.Atoi(sizeYMatches[i][1])
			}

			seriesInfos = append(seriesInfos, SeriesInfo{
				T: sizeT,
				C: sizeC,
				Z: sizeZ,
				X: sizeX,
				Y: sizeY,
			})
		}
	}

	// Extract laser wavelength information
	//laserWavelengths := map[string]string{}
	//for i, match := range laserMatches {
	//	channelKey := fmt.Sprintf("Channel %d", i+1)
	//	laserWavelengths[channelKey] = match[1]
	//}

	// Extract DateAndTime
	dateTimeRegex := regexp.MustCompile(`DateAndTime\s*:\s*(\S+\s*\S+)`)
	dateTimeMatch := dateTimeRegex.FindStringSubmatch(outputStr)
	dateAndTime := ""
	if len(dateTimeMatch) > 1 {
		dateAndTime = dateTimeMatch[1]
	}

	// Create the Metadata struct and return it
	metadata := &Metadata{
		SeriesInfos: seriesInfos,
		//LaserWavelengths: laserWavelengths,
		DateAndTime: dateAndTime,
	}

	return metadata, nil
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
	metadata, err := readImageMetadata(*imagePath, *jarPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Print metadata
	fmt.Printf("metadata: %v\n", metadata)

}
