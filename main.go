package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"

	"github.com/oodegard/parse_ims_metadata_txt"
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

// SampleInfo struct for user input fields
type SampleInfo struct {
	Channels   map[string]ChannelSample `yaml:"Channels"`
	ELNID      string                   `yaml:"ELNID"`
	SampleName string                   `yaml:"Sample name"`
}

// ChannelSample struct for user-defined channel details
type ChannelSample struct {
	Fluorophore string `yaml:"Fluorophore"`
}

// FinalYAML struct to hold the complete structure for YAML output
type FinalYAML struct {
	AcquisitionMetadata AcquisitionMetadata `yaml:"Acquisition metadata"`
	SampleInfo          SampleInfo          `yaml:"Sample info"`
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
	EmissionWavelength   float64 `yaml:"Emission wavelength"`
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

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func parseFlags() (path, ext string, showHelp bool) {
	pathPtr := flag.String("path", "", "Path to the image file or path to a directory containing images")
	extPtr := flag.String("ext", "", "File extension to filter images") // obligatory if image path is a directory
	helpPtr := flag.Bool("h", false, "Display help")

	flag.Parse()

	return *pathPtr, *extPtr, *helpPtr
}

func validateFlags(path, ext string, showHelp bool) {
	if showHelp {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if path == "" {
		log.Fatal("Image path is required. Use -path flag to specify the path.")
	}
}

func processFolder(path, ext, jarPath string) {
	sampleInfo, err := initializeSampleInfo(path, ext, jarPath)
	if err != nil {
		log.Fatalf("Error initializing sample info: %v", err)
	}

	processDirectory(path, ext, jarPath, sampleInfo)
}

func initializeSampleInfo(path, ext, jarPath string) (*SampleInfo, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %v", err)
	}

	firstImage := ""
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ext) {
			firstImage = filepath.Join(path, file.Name())
			break
		}
	}

	if firstImage == "" {
		return nil, fmt.Errorf("no image files found in the directory with the specified extension")
	}

	metadata, err := bfReadImageMetadata(firstImage, jarPath)
	if err != nil {
		return nil, fmt.Errorf("error reading metadata from the first image: %v", err)
	}

	numChannels := len(metadata.Images[0].Pixels.Channels)
	sampleInfoFile := filepath.Join(path, "inputSampleInfo.yaml")

	var sampleInfo *SampleInfo
	if _, err := os.Stat(sampleInfoFile); err == nil {
		fmt.Printf("Sample info file %s already exists.\n", sampleInfoFile)
		if waitForUserConfirmation() {
			yamlData, err := os.ReadFile(sampleInfoFile)
			if err != nil {
				return nil, fmt.Errorf("error reading sample info file: %v", err)
			}

			err = yaml.Unmarshal(yamlData, &sampleInfo)
			if err != nil {
				return nil, fmt.Errorf("error unmarshaling sample info YAML: %v", err)
			}
		} else {
			err := os.Remove(sampleInfoFile)
			if err != nil {
				return nil, fmt.Errorf("error deleting sample info file: %v", err)
			}

			err = createSampleInfoFile(sampleInfoFile, numChannels)
			if err != nil {
				return nil, fmt.Errorf("error creating sample info file: %v", err)
			}

			sampleInfo = createSampleInfo(numChannels)
		}
	} else {
		err = createSampleInfoFile(sampleInfoFile, numChannels)
		if err != nil {
			return nil, fmt.Errorf("error creating sample info file: %v", err)
		}

		if waitForUserConfirmation() {
			yamlData, err := os.ReadFile(sampleInfoFile)
			if err != nil {
				return nil, fmt.Errorf("error reading sample info file: %v", err)
			}

			err = yaml.Unmarshal(yamlData, &sampleInfo)
			if err != nil {
				return nil, fmt.Errorf("error unmarshaling sample info YAML: %v", err)
			}
		} else {
			err := os.Remove(sampleInfoFile)
			if err != nil {
				return nil, fmt.Errorf("error deleting sample info file: %v", err)
			}

			sampleInfo = createSampleInfo(numChannels)
		}
	}

	return sampleInfo, nil
}

func createSampleInfo(numChannels int) *SampleInfo {
	sampleInfo := &SampleInfo{
		Channels:   make(map[string]ChannelSample),
		ELNID:      "Please_fill_in_ELN_ID",
		SampleName: "Please_fill_in_a_sample_name",
	}

	for i := 1; i <= numChannels; i++ {
		if i <= 2 {
			sampleInfo.Channels[fmt.Sprintf("Channel %d", i)] = ChannelSample{
				Fluorophore: "Please_fill_in_a_fluorophore",
			}
		} else {
			sampleInfo.Channels[fmt.Sprintf("Channel %d", i)] = ChannelSample{
				Fluorophore: "NA",
			}
		}
	}

	return sampleInfo
}

func processImage(imagePath, jarPath string, sampleInfo *SampleInfo) error {
	// if imagePath ends with .ims
	if filepath.Ext(imagePath) == ".ims" {
		var metadataPath string

		// Check if it matches the pattern _F[number].ims
		if matched, _ := regexp.MatchString(`_F\d+\.ims$`, imagePath); matched {
			// Always replace the pattern with _metadata.txt
			base := strings.TrimSuffix(imagePath, filepath.Ext(imagePath))
			base = regexp.MustCompile(`_F\d+$`).ReplaceAllString(base, "")
			metadataPath = base + "_metadata.txt"
		} else {
			// Otherwise, just replace the .ims extension with _metadata.txt
			metadataPath = strings.TrimSuffix(imagePath, ".ims") + "_metadata.txt"
		}

		// Check if the metadata file exists
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			fmt.Printf("Metadata file does not exist for %s\n", imagePath)
			return nil
		}

		// Read metadata file
		ims_metadata := parse_ims_metadata_txt.ParseImsMetadatatxt(metadataPath)
		fmt.Printf("ims_metadata: %v\n", ims_metadata)

		// if it did not return already then try bfReadImageMetadata (below)
	}

	metadata, err := bfReadImageMetadata(imagePath, jarPath)
	if err != nil {
		log.Fatalf("Skipping file %s: %v", imagePath, err)
		return nil // Continue processing other files
	}

	metadataOutPath := strings.TrimSuffix(imagePath, filepath.Ext(imagePath)) + "_sampleInfo.yaml"
	err = saveMetadataAsYAML(metadata, metadataOutPath, sampleInfo)
	if err != nil {
		log.Printf("Error saving metadata as YAML for file %s: %v", imagePath, err)
		return nil // Continue processing other files
	}
	return nil
}

func processDirectory(path, ext, jarPath string, sampleInfo *SampleInfo) {
	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ext) {
			filePath := filepath.Join(path, file.Name())
			err := processImage(filePath, jarPath, sampleInfo)
			if err != nil {
				log.Printf("Error processing file %s: %v", filePath, err)
			}
		}
	}
}

func bfReadImageMetadata(imagePath, jarPath string) (*Metadata, error) {
	cmd := exec.Command("java", "-Dfile.encoding=UTF-8", "-cp", jarPath, "loci.formats.tools.ImageInfo", "-omexml-only", "-nopix", imagePath)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error parsing XML: %v", err)
	}

	output := out.String()
	var metadata Metadata
	err = xml.Unmarshal([]byte(output), &metadata)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling XML: %v", err)
	}

	return &metadata, nil
}

func saveMetadataAsYAML(metadata *Metadata, metadataOutPath string, sampleInfo *SampleInfo) error {
	acquisition := AcquisitionMetadata{
		Channels:    make(map[string]ChannelInfo),
		DateAndTime: metadata.Images[0].AcquisitionDate,
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
			PhysicalSizeZ: ptr(metadata.Images[0].Pixels.PhysicalSizeZ),
		},
	}

	for i, channel := range metadata.Images[0].Pixels.Channels {
		channelInfo := ChannelInfo{
			NameID:               fmt.Sprintf("%s %s", channel.ID, channel.Name),
			EmissionWavelength:   channel.EmissionWavelength,
			ExcitationWavelength: channel.ExcitationWavelength,
		}
		acquisition.Channels[fmt.Sprintf("Channel %d", i+1)] = channelInfo
	}

	finalYAML := FinalYAML{
		AcquisitionMetadata: acquisition,
		SampleInfo:          *sampleInfo,
	}

	yamlData, err := yaml.Marshal(finalYAML)
	if err != nil {
		return fmt.Errorf("error marshaling to YAML: %v", err)
	}

	err = os.WriteFile(metadataOutPath, yamlData, 0644)
	if err != nil {
		return fmt.Errorf("error writing to YAML file: %v", err)
	}

	fmt.Printf("YAML metadata saved to %s\n", metadataOutPath)
	return nil
}

func ptr(v float64) *float64 {
	return &v
}

func createSampleInfoFile(filePath string, numChannels int) error {
	sample := createSampleInfo(numChannels)

	yamlData, err := yaml.Marshal(sample)
	if err != nil {
		return fmt.Errorf("error marshaling sample info to YAML: %v", err)
	}

	err = os.WriteFile(filePath, yamlData, 0644)
	if err != nil {
		return fmt.Errorf("error writing sample info to YAML file: %v", err)
	}

	fmt.Printf("Sample info file created at %s. Please fill in the required values and then press 'y' to continue or 'n' to continue with default values.\n", filePath)
	return nil
}

func waitForUserConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Have you completed filling in the sample info file? Press 'y' to proceed or 'n' to continue with default values: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "y" || input == "yes" {
			return true
		}
		if input == "n" || input == "no" {
			return false
		}
	}
}

func main() {
	loadEnv()
	jarPath := os.Getenv("JAR_PATH")

	path, ext, showHelp := parseFlags()
	validateFlags(path, ext, showHelp)

	info, err := os.Stat(path)
	if err != nil {
		log.Fatalf("Error accessing path: %v", err)
	}

	if info.IsDir() {
		if ext == "" {
			log.Fatal("File extension is required when providing a directory path.")
		}

		processFolder(path, ext, jarPath)
	} else {
		err := processImage(path, jarPath, nil)
		if err != nil {
			log.Printf("Error processing file %s: %v", path, err)
		}
	}
}
