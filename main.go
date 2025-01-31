package main

import (
	"fmt"
	"log"

	"gogetbfmetadata/bfmetadata"
)

func main() {

	// Call the function to print the help info
	helpMessage, err := bfmetadata.PrintHelp()
	if err != nil {
		log.Fatalf("Error printing bfconvert help: %v", err)
	}

	fmt.Println("bfconvert Help Message:")
	fmt.Println(helpMessage)

}
