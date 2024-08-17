package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

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

		bytesDownloaded, err := downloadEpisode(ctx, client, filepath, episode.Enclosure.URL)
		if err != nil {
			if bytesDownloaded > 0 {
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

func downloadEpisode(ctx context.Context, client *http.Client, filepath string, url string) (uint64, error) {
	file, err := os.Create(filepath)
	if err != nil {
		return 0, fmt.Errorf("failed to open file for output: %w", err)
	}

	byteCounter := &ByteCounter{}

	err = client.Download(ctx, url, file, byteCounter)
	if err != nil {
		file.Close()

		rmErr := os.Remove(filepath)
		if rmErr != nil {
			return byteCounter.TotalBytes, fmt.Errorf("failed to delete file: %w, failed to download episode: %w", rmErr, err)
		}

		return byteCounter.TotalBytes, fmt.Errorf("failed to download episode: %w", err)
	}

	return byteCounter.TotalBytes, file.Close()
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

		err = downloadEpisodeQuietly(ctx, client, filepath, episode.Enclosure.URL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			return 1
		}
	}

	return 0
}

func downloadEpisodeQuietly(ctx context.Context, client *http.Client, filepath string, url string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file for output: %w", err)
	}

	err = client.Download(ctx, url, file, io.Discard)
	if err != nil {
		file.Close()

		rmErr := os.Remove(filepath)
		if rmErr != nil {
			return fmt.Errorf("failed to delete file: %w, failed to download episode: %w", rmErr, err)
		}

		return fmt.Errorf("failed to download episode: %w", err)
	}

	return file.Close()
}
