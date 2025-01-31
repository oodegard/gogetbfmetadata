# Go Get BFMetadata

**Note: This project is currently under development and is not yet functional.**

## Overview

The aim of this project is to build a Go-based application to extract essential metadata from microscopy images using the `bfconvert` tool. This tool is part of the Bio-Formats suite used extensively in the life sciences for image conversion and metadata handling.

## Features

- **Go Language**: The application is developed in Go, leveraging its powerful concurrency model and robust standard library.
- **Platform Specific**: This project is designed and tested only for Windows environments.

## Prerequisites

- **Go**: Make sure Go is installed on your system. You can download it from the [official Go website](https://golang.org/).
- **Git**: Ensure Git is installed to fetch Go modules. You can download it from [git-scm.com](https://git-scm.com/).
- **bfconvert**: Part of the Bio-Formats toolset, usually installed alongside the Open Microscopy Environment (OME).

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/oodegard/gogetbfmetadata.git
   ```

2. Navigate to the project directory:

   ```bash
   cd gogetbfmetadata
   ```

3. Ensure Go modules are correctly set up:

   ```bash
   go mod tidy
   ```

## Usage

Since the project is currently under development, the functionality is limited. The primary interaction involves trying to extract help information from the `bfconvert` tool for educational and scripting purposes.

To run the current implementation:

1. Navigate to the main directory and execute the main script:

   ```bash
   go run main.go
   ```

This will attempt to display help information from the included `bfconvert` executable.

## Project Structure

- `bfmetadata/`: Contains package code responsible for interacting with bfconvert.
- `bftools/`: Includes the `bfconvert` executable needed for operations.
- `main.go`: Entry point for executing the current implementation.

## Future Plans

This project aims to develop capabilities beyond simple interaction with `bfconvert`. Future iterations will include:

- Parsing and handling specific metadata from a variety of microscopy image formats.
- Providing a user-friendly interface for navigating extracted metadata.

## Contributions

Since this is an under-development project, contributions are welcome for educational and developmental guidance.

## Developer Info

Authored by **Øyvind Ødegård fougner**, currently focused on extending Go's capabilities into the realm of scientific data handling on Windows.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

Please note that any use of this repository represents acceptance of its current unfinished state and is engaged with at your own risk.
```

### Key Points Included:

1. **Development Status**: Clearly states that the project is under development and not yet operational.

2. **Goals and Purpose**: Outlines the primary aim and technical focus of the project.

3. **Dependencies and Setup**: Specifies necessary prerequisites and setup procedures, such as installing Go and Git.

4. **Usage Instructions**: Provides a simple usage guide aligned with the current implementation status.

5. **Future Directions**: Suggests future plans and extensions that will be part of the project.

6. **Contribution and Licensing**: Offers a note on potential contributions and includes a placeholder note for licensing.

This `README.md` serves both as technical documentation and a status update communication tool for anyone interested in the project's purpose and ongoing development.
