package bfmetadata

import (
	"bytes"
	_ "embed"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type OME struct {
	XMLName xml.Name `xml:"OME"`
	Image   Image    `xml:"Image"`
}

type Image struct {
	ID              string `xml:"ID,attr"`
	Name            string `xml:"Name,attr"`
	AcquisitionDate string `xml:"AcquisitionDate"`
	Pixels          Pixels `xml:"Pixels"`
}

type Pixels struct {
	BigEndian         string `xml:"BigEndian,attr"`
	DimensionOrder    string `xml:"DimensionOrder,attr"`
	ID                string `xml:"ID,attr"`
	Interleaved       string `xml:"Interleaved,attr"`
	PhysicalSizeX     string `xml:"PhysicalSizeX,attr"`
	PhysicalSizeXUnit string `xml:"PhysicalSizeXUnit,attr"`
	PhysicalSizeY     string `xml:"PhysicalSizeY,attr"`
	PhysicalSizeYUnit string `xml:"PhysicalSizeYUnit,attr"`
	PhysicalSizeZ     string `xml:"PhysicalSizeZ,attr"`
	PhysicalSizeZUnit string `xml:"PhysicalSizeZUnit,attr"`
	SignificantBits   int    `xml:"SignificantBits,attr"`
	SizeC             int    `xml:"SizeC,attr"`
	SizeT             int    `xml:"SizeT,attr"`
	SizeX             int    `xml:"SizeX,attr"`
	SizeY             int    `xml:"SizeY,attr"`
	SizeZ             int    `xml:"SizeZ,attr"`
	Type              string `xml:"Type,attr"`
}

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

// Embed showinf.bat or configure it if necessary
//
//go:embed bftools/showinf.bat
var showinfBat []byte

// PrintHelp executes the bfconvert.bat with the --help flag and returns the output.
func PrintHelp() (string, error) {
	tempDir, err := prepareFiles()
	if err != nil {
		return "", err
	}

	batFile := filepath.Join(tempDir, "bfconvert.bat")

	cmd := exec.Command("cmd", "/C", batFile, "--help")
	cmd.Env = append(os.Environ(), fmt.Sprintf("BF_DIR=%s", tempDir))

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return out.String(), fmt.Errorf("error executing bfconvert.bat --help: %w, raw stderr: %s", err, stderr.String())
	}

	return out.String(), nil
}

// GetOmexmlMetadata extracts and cleans OME-XML metadata from a given file using showinf.bat
func GetOmexmlMetadata(filePath string) (string, error) {
	tempDir, err := prepareFiles()
	if err != nil {
		return "", err
	}

	batFile := filepath.Join(tempDir, "showinf.bat")

	// Prepare the command to execute showinf.bat with -nopix to extract metadata
	cmd := exec.Command("cmd", "/C", batFile, filePath, "-omexml-only", "-nopix")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	cmd.Env = append(os.Environ(), fmt.Sprintf("BF_DIR=%s", tempDir))

	// Execute the command
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error executing showinf.bat to get metadata: %w, stderr: %s", err, stderr.String())
	}

	// Capture the command output and clean it to extract the XML content
	output := out.String()
	xmlIndex := strings.Index(output, "<?xml")
	if xmlIndex != -1 {
		return output[xmlIndex:], nil
	}

	return "", fmt.Errorf("no XML content found in output: %s", stderr.String())
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
		"showinf.bat":            showinfBat,
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

func GetEssentialMetadata(imageFilePath string) (map[string]interface{}, error) {
	// Simulate retrieving OME-XML metadata for the specified image file
	metadataxml, err := GetOmexmlMetadata(imageFilePath)
	if err != nil {
		return nil, err
	}

	metadata, err := parseXML(metadataxml)
	if err != nil {
		return nil, err
	}

	// Organize data into a format suitable for YAML
	essentialMetadata := map[string]interface{}{
		"Essential_metadata": map[string]interface{}{
			"AcquisitionDate": metadata.Image.AcquisitionDate,
			"DimensionOrder":  metadata.Image.Pixels.DimensionOrder,
			"PhysicalSize": map[string]interface{}{
				"X": metadata.Image.Pixels.PhysicalSizeX + " " + metadata.Image.Pixels.PhysicalSizeXUnit,
				"Y": metadata.Image.Pixels.PhysicalSizeY + " " + metadata.Image.Pixels.PhysicalSizeYUnit,
				"Z": metadata.Image.Pixels.PhysicalSizeZ + " " + metadata.Image.Pixels.PhysicalSizeZUnit,
			},
			"Size": map[string]interface{}{
				"C": metadata.Image.Pixels.SizeC,
				"T": metadata.Image.Pixels.SizeT,
				"X": metadata.Image.Pixels.SizeX,
				"Y": metadata.Image.Pixels.SizeY,
				"Z": metadata.Image.Pixels.SizeZ,
			},
			"PixelBitDepth": metadata.Image.Pixels.SignificantBits,
		},
	}

	return essentialMetadata, nil
}

func parseXML(xmlData string) (*OME, error) {
	var ome OME

	reader := strings.NewReader(xmlData)
	decoder := xml.NewDecoder(reader)
	decoder.DefaultSpace = ""

	if err := decoder.Decode(&ome); err != nil {
		return nil, err
	}

	return &ome, nil
}
