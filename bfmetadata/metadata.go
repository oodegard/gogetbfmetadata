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

// Node represents a generic XML element structure
type Node struct {
	XMLName xml.Name
	Content []byte `xml:",innerxml"`
	Nodes   []Node `xml:",any"`
}

// Define structs according to the XML schema
type Instrument struct {
	ID        string    `xml:"ID,attr"`
	Detector  Detector  `xml:"Detector"`
	Objective Objective `xml:"Objective"`
}

type Detector struct {
	ID    string `xml:"ID,attr"`
	Model string `xml:"Model,attr"`
	Type  string `xml:"Type,attr"`
}

type Objective struct {
	ID                   string `xml:"ID,attr"`
	Correction           string `xml:"Correction,attr"`
	Immersion            string `xml:"Immersion,attr"`
	LensNA               string `xml:"LensNA,attr"`
	Manufacturer         string `xml:"Manufacturer,attr"`
	Model                string `xml:"Model,attr"`
	NominalMagnification string `xml:"NominalMagnification,attr"`
	WorkingDistance      string `xml:"WorkingDistance,attr"`
	WorkingDistanceUnit  string `xml:"WorkingDistanceUnit,attr"`
}

type Image struct {
	ID              string `xml:"ID,attr"`
	Name            string `xml:"Name,attr"`
	AcquisitionDate string `xml:"AcquisitionDate"`
	InstrumentRef   Ref    `xml:"InstrumentRef"`
	ObjectiveRef    Ref    `xml:"ObjectiveSettings"`
	Pixels          Pixels `xml:"Pixels"`
}

type Ref struct {
	ID string `xml:"ID,attr"`
}

type Pixels struct {
	BigEndian      string    `xml:"BigEndian,attr"`
	DimensionOrder string    `xml:"DimensionOrder,attr"`
	ID             string    `xml:"ID,attr"`
	Interleaved    string    `xml:"Interleaved,attr"`
	PhysicalSizeX  string    `xml:"PhysicalSizeX,attr"`
	PhysicalSizeY  string    `xml:"PhysicalSizeY,attr"`
	PhysicalSizeZ  string    `xml:"PhysicalSizeZ,attr"`
	SizeC          string    `xml:"SizeC,attr"`
	SizeT          string    `xml:"SizeT,attr"`
	SizeX          string    `xml:"SizeX,attr"`
	SizeY          string    `xml:"SizeY,attr"`
	SizeZ          string    `xml:"SizeZ,attr"`
	Type           string    `xml:"Type,attr"`
	Channels       []Channel `xml:"Channel"`
}

type Channel struct {
	EmissionWavelength   string `xml:"EmissionWavelength,attr"`
	ExcitationWavelength string `xml:"ExcitationWavelength,attr"`
	ID                   string `xml:"ID,attr"`
	Name                 string `xml:"Name,attr"`
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

// GetMetadata retrieves and processes specific XML metadata
func GetMetadata(filePath string) (Instrument, Image, error) {
	metadataXML, err := GetOmexmlMetadata(filePath)
	if err != nil {
		return Instrument{}, Image{}, fmt.Errorf("error retrieving OME-XML metadata: %w", err)
	}

	var instr Instrument
	var img Image

	err = ParseAndProcessMetadata(metadataXML, func(node Node) bool {
		switch node.XMLName.Local {
		case "Instrument":
			if err := xml.Unmarshal(node.Content, &instr); err != nil {
				fmt.Printf("Error unmarshalling Instrument: %v\n", err)
			}
		case "Image":
			if err := xml.Unmarshal(node.Content, &img); err != nil {
				fmt.Printf("Error unmarshalling Image: %v\n", err)
			}
		}
		return true
	})
	if err != nil {
		return Instrument{}, Image{}, fmt.Errorf("error processing XML metadata: %w", err)
	}

	return instr, img, nil
}

// Efficiently parse and process the metadata XML using Node structure
func ParseAndProcessMetadata(xmlData string, processFunc func(Node) bool) error {
	var root Node
	if err := xml.Unmarshal([]byte(xmlData), &root); err != nil {
		return fmt.Errorf("error parsing XML: %w", err)
	}

	walk(root.Nodes, processFunc)
	return nil
}

// Recursive function to walk through nodes
func walk(nodes []Node, processFunc func(Node) bool) {
	for _, n := range nodes {
		if processFunc(n) {
			walk(n.Nodes, processFunc)
		}
	}
}
