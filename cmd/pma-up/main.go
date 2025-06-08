package main

import (
	"log"
	"os"

	"github.com/jsas4coding/pma-up/internal/updater"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: pma-up <destination_path> <config_file_path>")
	}

	destinationPath := os.Args[1]
	configFilePath := os.Args[2]

	if err := updater.RunUpdate(destinationPath, configFilePath); err != nil {
		log.Fatalf("Update failed: %v", err)
	}
}
