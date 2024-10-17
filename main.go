package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

// Metadata struct to hold the entire XML structure
type Metadata struct {
	XMLName xml.Name `xml:"OME"`
	Images  []Image  `xml:"Image"`
}

// Image struct to hold image-specific metadata
type Image struct {
	ID              string `xml:"ID,attr"`
	Name            string `xml:"Name,attr"`
	AcquisitionDate string `xml:"AcquisitionDate"`
	Pixels          Pixels `xml:"Pixels"`
}

// Pixels struct to hold pixel-specific metadata
type Pixels struct {
	SizeX         int       `xml:"SizeX,attr"`
	SizeY         int       `xml:"SizeY,attr"`
	SizeZ         int       `xml:"SizeZ,attr"`
	SizeC         int       `xml:"SizeC,attr"`
	SizeT         int       `xml:"SizeT,attr"`
	PhysicalSizeX float64   `xml:"PhysicalSizeX,attr"`
	PhysicalSizeY float64   `xml:"PhysicalSizeY,attr"`
	PhysicalSizeZ float64   `xml:"PhysicalSizeZ,attr"`
	Channels      []Channel `xml:"Channel"` // Slice to hold channels
}

// Channel struct to hold channel-specific metadata
type Channel struct {
	ID                   string  `xml:"ID,attr"`
	Name                 string  `xml:"Name,attr"`
	EmissionWavelength   float64 `xml:"EmissionWavelength,attr"`
	ExcitationWavelength float64 `xml:"ExcitationWavelength,attr"`
	SamplesPerPixel      int     `xml:"SamplesPerPixel,attr"`
}

// AcquisitionMetadata struct to match desired YAML structure
type AcquisitionMetadata struct {
	Channels    map[string]ChannelInfo `yaml:"Channels"`
	DateAndTime string                 `yaml:"DateAndTime"`
	ImageDims   ImageDimensions        `yaml:"Image dimensions"`
	PixelSize   PixelSizeInfo          `yaml:"Pixel size"`
}

// ChannelInfo struct for channel details in YAML
type ChannelInfo struct {
	NameID               string  `yaml:"Name ID"`
	EmissionWavelength   float64 `yaml:"Emmision wavelength"`
	ExcitationWavelength float64 `yaml:"Excitation wavelength"`
}

// ImageDimensions struct for image dimensions in YAML
type ImageDimensions struct {
	SizeC int `yaml:"SizeC"`
	SizeT int `yaml:"SizeT"`
	SizeX int `yaml:"SizeX"`
	SizeY int `yaml:"SizeY"`
	SizeZ int `yaml:"SizeZ"`
}

// PixelSizeInfo struct for pixel size information in YAML
type PixelSizeInfo struct {
	PhysicalSizeX float64  `yaml:"PhysicalSizeX"`
	PhysicalSizeY float64  `yaml:"PhysicalSizeY"`
	PhysicalSizeZ *float64 `yaml:"PhysicalSizeZ"`
}

// SampleInfo struct for user input fields
type SampleInfo struct {
	Channels map[string]ChannelSample `yaml:"Channels"`
	ELNID    string                   `yaml:"ELNID"`
}

// ChannelSample struct for user-defined channel details
type ChannelSample struct {
	Fluorophore string `yaml:"Fluorophore"`
	SampleName  string `yaml:"Sample name"`
}

// FinalYAML struct to hold the complete structure for YAML output
type FinalYAML struct {
	AquisitionMetadata AcquisitionMetadata `yaml:"Aquisition metadata"`
	SampleInfo         SampleInfo          `yaml:"Sample info"`
}

// Function to read metadata from an image using Bio-Formats
func readImageMetadata(imagePath, jarPath string) (*Metadata, error) {
	cmd := exec.Command("java", "-Dfile.encoding=UTF-8", "-cp", jarPath, "loci.formats.tools.ImageInfo", "-omexml-only", "-nopix", imagePath)
	fmt.Printf("cmd: %v\n", cmd)
	// Capture standard output
	var out bytes.Buffer
	cmd.Stdout = &out

	// Run the command
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return nil, fmt.Errorf("error parsing XML: %v", err)
	}

	output := out.String()
	//fmt.Printf("output: %v\n", output)

	// Unmarshal the XML output into the Metadata struct
	var metadata Metadata
	err = xml.Unmarshal([]byte(output), &metadata)

	// We are not interested in the thumbnail images so we remove them from the metadata
	if err != nil {
		fmt.Println("Error:", err)
		return nil, fmt.Errorf("error Unmarshaling XML: %v", err)
	}

	return &metadata, nil
}

// Function to create YAML structure and save it to a file
// Function to create YAML structure and save it to a file
func saveMetadataAsYAML(metadata *Metadata, metadataOutPath string) error {
	// Prepare the acquisition metadata structure
	acquisition := AcquisitionMetadata{
		Channels:    make(map[string]ChannelInfo),
		DateAndTime: metadata.Images[0].AcquisitionDate, // Placeholder for date and time
		ImageDims: ImageDimensions{
			SizeC: metadata.Images[0].Pixels.SizeC,
			SizeT: metadata.Images[0].Pixels.SizeT,
			SizeX: metadata.Images[0].Pixels.SizeX,
			SizeY: metadata.Images[0].Pixels.SizeY,
			SizeZ: metadata.Images[0].Pixels.SizeZ,
		},
		PixelSize: PixelSizeInfo{
			PhysicalSizeX: metadata.Images[0].Pixels.PhysicalSizeX,
			PhysicalSizeY: metadata.Images[0].Pixels.PhysicalSizeY,
			PhysicalSizeZ: nil, // Set to nil or appropriate value
		},
	}

	// Fill in channel information for each channel in the metadata
	for i, channel := range metadata.Images[0].Pixels.Channels {
		channelInfo := ChannelInfo{
			NameID:               fmt.Sprintf("%s %s", channel.ID, channel.Name),
			EmissionWavelength:   channel.EmissionWavelength,
			ExcitationWavelength: channel.ExcitationWavelength,
		}
		acquisition.Channels[fmt.Sprintf("Channel %d", i+1)] = channelInfo // Use i+1 to start numbering from 1
	}

	// Prepare the sample information structure with placeholders
	sample := SampleInfo{
		Channels: make(map[string]ChannelSample),
		ELNID:    "Please_fill_in_ELN_ID",
	}

	// Create placeholders for each channel
	for i, _ := range metadata.Images[0].Pixels.Channels {
		channelKey := fmt.Sprintf("Channel %d", i+1)
		sample.Channels[channelKey] = ChannelSample{
			Fluorophore: "Please_fill_in_a_fluorophore",
			SampleName:  "Please_fill_in_a_sample_name",
		}
	}

	// Combine acquisition metadata and sample information
	finalYAML := FinalYAML{
		AquisitionMetadata: acquisition,
		SampleInfo:         sample,
	}

	// Convert to YAML
	yamlData, err := yaml.Marshal(finalYAML)
	if err != nil {
		return fmt.Errorf("error marshaling to YAML: %v", err)
	}

	// Write YAML data to file
	err = os.WriteFile(metadataOutPath, yamlData, 0644)
	if err != nil {
		return fmt.Errorf("error writing to YAML file: %v", err)
	}

	fmt.Printf("YAML metadata saved to %s\n", metadataOutPath)
	return nil
}

// Main function
func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get the jarPath variable
	jarPath := os.Getenv("JAR_PATH")

	// Define flags
	imagePathPtr := flag.String("image", "", "Path to the image file")
	helpPtr := flag.Bool("h", false, "Display help")

	// Parse flags
	flag.Parse()

	// Show help message if requested
	if *helpPtr {
		flag.PrintDefaults()
		return
	}

	// Check if image path is provided
	if *imagePathPtr == "" {
		log.Fatal("Image path is required. Use -image flag to specify the path.")
	}

	// Read the metadata from the image file
	metadata, err := readImageMetadata(*imagePathPtr, jarPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("metadata: %v\n", metadata)

	// replace file extension with "_sampleInfo.yaml"
	metadataOutPath := strings.TrimSuffix(*imagePathPtr, filepath.Ext(*imagePathPtr)) + "_sampleInfo.yaml"
	// Save metadata to YAML file
	err = saveMetadataAsYAML(metadata, metadataOutPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
