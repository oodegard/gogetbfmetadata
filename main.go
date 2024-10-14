package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

// Struct to hold series information
type SeriesInfo struct {
	Timepoints int
	Channels   int
	ZStacks    int
}

// Function to check if Java is installed
func checkJavaInstallation() error {
	_, err := exec.LookPath("java")
	if err != nil {
		return fmt.Errorf("Java is not installed or not in your PATH")
	}
	return nil
}

// Function to check if the series are reduced versions of each other (based on size)
func isReducedVersion(seriesInfos []SeriesInfo, widths []int, heights []int) bool {
	for i := 1; i < len(seriesInfos); i++ {
		// Check if width and height are halved in successive series
		if widths[i] != widths[i-1]/2 || heights[i] != heights[i-1]/2 {
			return false
		}
	}
	return true
}

// Function to read metadata from an image using Bio-Formats
func readImageMetadata(imagePath, jarPath string) ([]SeriesInfo, error) {
	// The "-nopix" flag tells Bio-Formats to read only the metadata, skipping pixel data
	cmd := exec.Command("java", "-cp", jarPath, "loci.formats.tools.ImageInfo", "-nopix", imagePath)

	// Run the command and capture the output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error running Java command: %v", err)
	}

	// Convert output to string for parsing
	outputStr := string(output)

	// Use regular expressions to extract relevant metadata
	seriesRegex := regexp.MustCompile(`Series count = (/d+)`)
	sizeTRegex := regexp.MustCompile(`SizeT = (/d+)`)
	sizeCRegex := regexp.MustCompile(`SizeC = (/d+)`)
	sizeZRegex := regexp.MustCompile(`SizeZ = (/d+)`)

	// Extract values
	seriesMatches := seriesRegex.FindStringSubmatch(outputStr)
	sizeTMatches := sizeTRegex.FindAllStringSubmatch(outputStr, -1)
	sizeCMatches := sizeCRegex.FindAllStringSubmatch(outputStr, -1)
	sizeZMatches := sizeZRegex.FindAllStringSubmatch(outputStr, -1)

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

		if i < len(sizeTMatches) && len(sizeTMatches[i]) > 1 {
			sizeT, _ = strconv.Atoi(sizeTMatches[i][1])
		}
		if i < len(sizeCMatches) && len(sizeCMatches[i]) > 1 {
			sizeC, _ = strconv.Atoi(sizeCMatches[i][1])
		}
		if i < len(sizeZMatches) && len(sizeZMatches[i]) > 1 {
			sizeZ, _ = strconv.Atoi(sizeZMatches[i][1])
		}

		seriesInfos = append(seriesInfos, SeriesInfo{
			Timepoints: sizeT,
			Channels:   sizeC,
			ZStacks:    sizeZ,
		})
	}

	return seriesInfos, nil
}

// Function to group and print series with the same metadata
func printSeriesGroups(seriesInfos []SeriesInfo) {
	n := len(seriesInfos)
	if n == 0 {
		fmt.Println("No series information available.")
		return
	}

	groupStart := 0
	for i := 1; i <= n; i++ {
		if i == n || seriesInfos[i] != seriesInfos[groupStart] {
			// If end of group is reached, print the group
			if groupStart == i-1 {
				fmt.Printf("Series #%d:/n", groupStart)
			} else {
				fmt.Printf("Series #%d-%d:/n", groupStart, i-1)
			}
			fmt.Printf("  Timepoints (T): %d/n", seriesInfos[groupStart].Timepoints)
			fmt.Printf("  Channels (C): %d/n", seriesInfos[groupStart].Channels)
			fmt.Printf("  Z-stacks (Z): %d/n", seriesInfos[groupStart].ZStacks)

			// Move groupStart to the current series
			groupStart = i
		}
	}
}

func main() {
	// Check if Java is installed
	if err := checkJavaInstallation(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Get command-line arguments for the image and jar paths
	imagePath := "./path/to/your/image.ims"
	imagePath = "C:/Users/Øyvind/OneDrive - Universitetet i Oslo/Work/03_UiO/01_ELN_auto/Barcoded nanobodies/127 - Barcoding all nanobodies + Halo dyes/DragonFly/20240321_abwo_IBIDI2/del/20240321_127_IBIDI2B_17_18_1.ims"
	jarPath := "./bioformats_package.jar"

	if len(os.Args) > 1 {
		imagePath = os.Args[1]
	}
	if len(os.Args) > 2 {
		jarPath = os.Args[2]
	}

	// Read the metadata from the image file
	seriesInfos, err := readImageMetadata(imagePath, jarPath)
	if err != nil {
		fmt.Printf("Error: %v/n", err)
		os.Exit(1)
	}

	// Group and print series information
	printSeriesGroups(seriesInfos)
}
