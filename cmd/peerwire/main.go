package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Minesto23/peerwire/internal/engine"
	"github.com/Minesto23/peerwire/internal/torrent"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: peerwire download <file.torrent> [output_path]")
		return
	}

	command := os.Args[1]

	if command == "download" {
		if len(os.Args) < 3 {
			fmt.Println("Usage: peerwire download <file.torrent> [output_path]")
			return
		}

		torrentPath := os.Args[2]
		outputPath := "."
		if len(os.Args) > 3 {
			outputPath = os.Args[3]
		}

		// 1. Parse Torrent
		f, err := os.Open(torrentPath)
		if err != nil {
			fmt.Printf("Error opening torrent file: %v\n", err)
			return
		}
		defer f.Close()

		spec, err := torrent.Parse(f)
		if err != nil {
			fmt.Printf("Error parsing torrent: %v\n", err)
			return
		}

		fmt.Printf("File: %s\nLength: %d bytes\n", spec.Info.Name, spec.Info.Length)

		targetFile := filepath.Join(outputPath, spec.Info.Name)

		// 2. Start Engine
		params := engine.ClientParams{
			OutputPath: targetFile,
		}

		client, err := engine.NewClient(spec, params)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			return
		}

		if err := client.Download(func(done, total int) {
			percent := float64(done) / float64(total) * 100
			fmt.Printf("\rDownloaded: %0.2f%% (%d/%d pieces)", percent, done, total)
		}); err != nil {
			fmt.Printf("\nDownload error: %v\n", err)
		} else {
			fmt.Println("\nDownload Complete!")
		}

	} else {
		fmt.Printf("Unknown command: %s\n", command)
	}
}
