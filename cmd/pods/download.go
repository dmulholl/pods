package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dmulholl/pods/internal/counter"
	"github.com/dmulholl/pods/internal/http"
	"github.com/dmulholl/pods/internal/rss"
	"github.com/dmulholl/pods/internal/term"
)

func downloadEpisodes(ctx context.Context, dstDirectory string, filenameFormat string, episodes []rss.Item) int {
	client := http.NewClient(5 * time.Minute)
	numErrors := 0

	for _, episode := range episodes {
		if numErrors > 5 {
			term.PrintRed("   Aborting")
			fmt.Printf(" too many errors\n")
			break
		}

		term.PrintGreen("Downloading")
		fmt.Printf(" [%d] %s\n", episode.Episode, episode.Title)

		filename, err := formatFilename(filenameFormat, episode)
		if err != nil {
			term.PrintRed("      Error")
			fmt.Printf(" failed to format filename: %s\n", err)
			numErrors += 1
			continue
		}

		filepath := filepath.Join(dstDirectory, filename)
		counter := counter.New(true)

		err = downloadEpisode(ctx, client, filepath, episode.Enclosure.URL, counter)
		if err != nil {
			if counter.TotalBytes() > 0 {
				fmt.Println()
			}

			term.PrintRed("      Error")
			fmt.Printf(" %s\n", err)

			numErrors += 1
			continue
		}

		fmt.Println(" [complete]")
	}

	return numErrors
}

func downloadEpisodesQuietly(ctx context.Context, dstDirectory string, filenameFormat string, episodes []rss.Item) int {
	client := http.NewClient(5 * time.Minute)

	for _, episode := range episodes {
		filename, err := formatFilename(filenameFormat, episode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to format filename: %s\n", err)
			return 1
		}

		filepath := filepath.Join(dstDirectory, filename)
		counter := counter.New(false)

		err = downloadEpisode(ctx, client, filepath, episode.Enclosure.URL, counter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			return 1
		}
	}

	return 0
}

func downloadEpisode(ctx context.Context, client *http.Client, filepath string, url string, counter *counter.ByteCounter) error {
	tempFilepath := filepath + ".temp"

	file, err := os.Create(tempFilepath)
	if err != nil {
		return fmt.Errorf("failed to open file for output: %w", err)
	}

	err = client.Download(ctx, url, file, counter)
	if err != nil {
		file.Close()
		os.Remove(tempFilepath)
		return fmt.Errorf("failed to download episode: %w", err)
	}

	err = file.Close()
	if err != nil {
		os.Remove(tempFilepath)
		return fmt.Errorf("failed to close file: %w", err)
	}

	err = os.Rename(tempFilepath, filepath)
	if err != nil {
		return fmt.Errorf("failed to rename downloaded file to remove '.temp' suffix: %w", err)
	}

	return nil
}
