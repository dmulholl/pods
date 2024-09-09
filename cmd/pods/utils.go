package main

import (
	"fmt"
	"mime"
	"strings"
	"time"

	"github.com/dmulholl/pods/internal/rss"
)

var extensions = map[string]string{
	"audio/mpeg": ".mp3",
	"audio/m4a":  ".m4a",
	"video/m4v":  ".m4v",
	"video/mp4":  ".mp4",
}

// Creates a filename from a template string containing '{{foo}}' placeholders.
func formatFilename(format string, episode rss.Item) (string, error) {
	filename := strings.Replace(format, "{{title}}", strings.TrimSpace(episode.Title), -1)

	if strings.Contains(filename, "{{ext}}") {
		ext, err := extensionForType(episode.Enclosure.Type)
		if err != nil {
			return "", fmt.Errorf("failed to determine the default file extension: %w", err)
		}

		filename = strings.Replace(filename, "{{ext}}", ext, -1)
	}

	filename = strings.Replace(filename, "{{episode}}", fmt.Sprintf("%d", episode.Episode), -1)
	filename = strings.Replace(filename, "{{episode2}}", fmt.Sprintf("%02d", episode.Episode), -1)
	filename = strings.Replace(filename, "{{episode3}}", fmt.Sprintf("%03d", episode.Episode), -1)
	filename = strings.Replace(filename, "{{episode4}}", fmt.Sprintf("%04d", episode.Episode), -1)

	filename = strings.Replace(filename, "{{season}}", fmt.Sprintf("%d", episode.Season), -1)
	filename = strings.Replace(filename, "{{season2}}", fmt.Sprintf("%02d", episode.Season), -1)
	filename = strings.Replace(filename, "{{season3}}", fmt.Sprintf("%03d", episode.Season), -1)
	filename = strings.Replace(filename, "{{season4}}", fmt.Sprintf("%04d", episode.Season), -1)

	return filename, nil
}

// Returns the file extension for the MIME type, e.g. '.mp3'.
func extensionForType(mimetype string) (string, error) {
	if ext, ok := extensions[mimetype]; ok {
		return ext, nil
	}

	extensions, err := mime.ExtensionsByType(mimetype)
	if err != nil {
		return "", fmt.Errorf("invalid MIME type '%s': %w", mimetype, err)
	}

	if len(extensions) == 0 {
		return "", fmt.Errorf("unknown MIME type: %s", mimetype)
	}

	return extensions[0], nil
}

// Parses an RSS publication date.
func parseRssPubDate(input string) (time.Time, error) {
	pubdate, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", input)
	if err != nil {
		pubdate, err = time.Parse("Mon, 02 Jan 2006 15:04:05 MST", input)
		if err != nil {
			return time.Time{}, err
		}
	}

	return pubdate, nil
}

// Parses a timestamp in RFC 3339 format.
//   - Supports shorter formats for convenience.
//   - If no timezone is specified, it's assumed to be UTC.
func parseInputTimestamp(input string) (time.Time, error) {
	dt, err := time.Parse("2006-01-02T15:04:05Z07:00", input) // RFC 3339
	if err != nil {
		dt, err = time.Parse("2006-01-02 15:04:05Z07:00", input) // RFC 3339
		if err != nil {
			dt, err = time.Parse("2006-01-02T15:04:05", input)
			if err != nil {
				dt, err = time.Parse("2006-01-02 15:04:05", input)
				if err != nil {
					dt, err = time.Parse("2006-01-02", input)
					if err != nil {
						return time.Time{}, err
					}
				}
			}
		}
	}

	return dt, nil
}
