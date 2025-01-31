package main

import (
	"fmt"
	"log"

	"gogetbfmetadata/bfmetadata"

	"github.com/joho/godotenv"
)

func main() {
	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Call the function to print the help info
	helpMessage, err := bfmetadata.PrintHelp()
	if err != nil {
		log.Fatalf("Error printing bfconvert help: %v", err)
	}

	fmt.Println("bfconvert Help Message:")
	fmt.Println(helpMessage)

}
