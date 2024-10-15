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
		return fmt.Errorf("java is not installed or not in your path: %v", err)
	}
	return nil
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
	//fmt.Printf("outputStr: %v\n", outputStr)

	// Use more flexible regular expressions to extract relevant metadata
	seriesRegex := regexp.MustCompile(`Series\s*count\s*=\s*(\d+)`)
	sizeTRegex := regexp.MustCompile(`SizeT\s*=\s*(\d+)`)
	sizeCRegex := regexp.MustCompile(`SizeC\s*=\s*(\d+)`)
	sizeZRegex := regexp.MustCompile(`SizeZ\s*=\s*(\d+)`)

	// Extract values
	seriesMatches := seriesRegex.FindStringSubmatch(outputStr)
	fmt.Printf("seriesMatches: %v\n", seriesMatches)
	sizeTMatches := sizeTRegex.FindAllStringSubmatch(outputStr, -1)
	sizeCMatches := sizeCRegex.FindAllStringSubmatch(outputStr, -1)
	sizeZMatches := sizeZRegex.FindAllStringSubmatch(outputStr, -1)
	fmt.Printf("sizeZMatches: %v\n", sizeZMatches)
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
				fmt.Printf("Series #%d:\n", groupStart)
			} else {
				fmt.Printf("Series #%d-%d:\n", groupStart, i-1)
			}
			fmt.Printf("  Timepoints (T): %d\n", seriesInfos[groupStart].Timepoints)
			fmt.Printf("  Channels (C): %d\n", seriesInfos[groupStart].Channels)
			fmt.Printf("  Z-stacks (Z): %d\n", seriesInfos[groupStart].ZStacks)

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
	imagePath := "C:/Users/Øyvind/OneDrive - Universitetet i Oslo/Work/03_UiO/04_Microscope_images_DO_NOT_USE/20210225_017/Phafin2-GFP_MAP4K3-Halo549/3_RPE-1_Phafin2-GFP_MAP4K3-Halo549_3sec_001_D3D_ALX.dv"
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
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Group and print series information
	printSeriesGroups(seriesInfos)
}
