package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/dmulholl/argo/v4"
	"github.com/dmulholl/pods/internal/http"
	"github.com/dmulholl/pods/internal/rss"
	"github.com/dmulholl/pods/internal/term"
)

const version = "v0.1.0"

var helptext = fmt.Sprintf(`
Pods %s

Usage:
  pods --url <rss-feed>

  A utility for downloading podcast episodes.

  By default, this utility simply lists the episodes which would be downloaded.
  Use the -d/--download flag to actually download the episodes.

  The --after/--before options accept a simple date, e.g.

    --after "2024-07-31"

  Or, a full RFC-3339 timestamp, e.g.

    --after "2024-07-31T13:59:00+02:00"

  If no timezone offset is specified, the timestamp is assumed to be UTC.

  The output filename can be customized using the -f/--format option. The
  following format specifiers are supported:

  - %%episode%%:  Episode number.
  - %%episode2%%: Episode number with zero-padding, min-width: 2 digits.
  - %%episode3%%: Episode number with zero-padding, min-width: 3 digits.
  - %%episode4%%: Episode number with zero-padding, min-width: 4 digits.
  - %%ext%%:      The default file extension for the file type, e.g. '.mp3'.
  - %%title%%:    The episode title.

  The default filename format is '%%episode4%%. %%title%%%%ext%%'.

Options:
  -a, --after <timestamp>   Download episodes published after this timestamp.
  -b, --before <timestamp>  Download episodes published before this timestamp.
  -e, --episode <number>    Download the specified episode number.
                            This option can be specified multiple times.
      --file <filepath>     Specifies a source file for the RSS feed.
  -f, --format <format>     Overrides the default format for output filenames.
                            Default: '%%episode4%%. %%title%%%%ext%%'.
  -o, --outdir <path>       Output directory for downloaded files.
                            Default: './<podcast-title>'.
  -u, --url <url>           Specifies a source URL for the RSS feed.

Flags:
  -d, --download            Download podcast episodes.
  -h, --help                Print the application's help text.
  -v, --version             Print the application's version number.
`, version)

var extensions = map[string]string{
	"audio/mpeg": ".mp3",
	"audio/m4a":  ".m4a",
	"video/m4v":  ".m4v",
	"video/mp4":  ".mp4",
}

func main() {
	argparser := argo.NewParser()
	argparser.Helptext = helptext
	argparser.Version = version

	argparser.NewStringOption("file", "")
	argparser.NewStringOption("url u", "")
	argparser.NewStringOption("before b", "")
	argparser.NewStringOption("after a", "")
	argparser.NewStringOption("outdir o", "")
	argparser.NewIntOption("episode e", 0)
	argparser.NewStringOption("format f", "%episode4%. %title%%ext%")
	argparser.NewFlag("download d")

	if err := argparser.ParseOsArgs(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	os.Exit(runMain(argparser))
}

func runMain(args *argo.ArgParser) int {
	ctx := context.Background()

	if !args.Found("file") && !args.Found("url") {
		fmt.Fprintf(os.Stderr, "error: expected --url or --file argument\n")
		return 1
	}

	var err error
	before := time.Time{}
	after := time.Time{}

	if args.Found("before") {
		before, err = parseInputTime(args.StringValue("before"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to parse 'before' date: '%s'\n", args.StringValue("before"))
			return 1
		}
	}

	if args.Found("after") {
		after, err = parseInputTime(args.StringValue("after"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to parse 'after' date: '%s\n", args.StringValue("after"))
			return 1
		}
	}

	var data []byte

	if args.Found("url") {
		client := http.NewClient(30 * time.Second)

		body, err := client.Get(ctx, args.StringValue("url"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to download RSS feed: %s\n", err)
			return 1
		}

		data = body
	}

	if args.Found("file") {
		content, err := os.ReadFile(args.StringValue("file"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to read RSS feed from file: %s\n", err)
			return 1
		}

		data = content
	}

	rssFeed := rss.RSS{}
	if err := xml.Unmarshal(data, &rssFeed); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to unmarshall RSS feed: %s\n", err)
		return 1
	}

	for _, channel := range rssFeed.Channels {
		episodes := []rss.Item{}

		for _, item := range channel.Items {
			pubdate, err := parseRssPubDate(item.PubDate)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: failed to parse publication date: '%s'\n", item.PubDate)
				return 1
			}

			if !before.IsZero() {
				if pubdate.After(before) {
					continue
				}
			}

			if !after.IsZero() {
				if pubdate.Before(after) {
					continue
				}
			}

			if args.Found("episode") {
				if !slices.Contains(args.IntValues("episode"), item.Episode) {
					continue
				}
			}

			episodes = append(episodes, item)
		}

		if !args.Found("download") {
			listEpisodes(channel.Title, episodes)
			continue
		}

		dstDirectory := args.StringValue("outdir")
		if dstDirectory == "" {
			dstDirectory = fmt.Sprintf("./%s", channel.Title)
		}

		err := os.MkdirAll(dstDirectory, 0o755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to create output directory: %s\n", err)
			return 1
		}

		numErrors := downloadEpisodes(ctx, dstDirectory, args.StringValue("format"), episodes)
		if numErrors > 0 {
			return 1
		}
	}

	return 0
}

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
		fmt.Printf(" %s\n", episode.Title)

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
		return byteCounter.TotalBytes, fmt.Errorf("failed to download episode: %w", err)
	}

	return byteCounter.TotalBytes, file.Close()
}

func formatFilename(format string, episode rss.Item) (string, error) {
	filename := strings.Replace(format, "%title%", strings.TrimSpace(episode.Title), -1)

	if strings.Contains(filename, "%ext%") {
		ext, err := extensionForType(episode.Enclosure.Type)
		if err != nil {
			return "", fmt.Errorf("failed to determine the default file extension: %w", err)
		}

		filename = strings.Replace(filename, "%ext%", ext, -1)
	}

	filename = strings.Replace(filename, "%episode%", fmt.Sprintf("%d", episode.Episode), -1)
	filename = strings.Replace(filename, "%episode2%", fmt.Sprintf("%02d", episode.Episode), -1)
	filename = strings.Replace(filename, "%episode3%", fmt.Sprintf("%03d", episode.Episode), -1)
	filename = strings.Replace(filename, "%episode4%", fmt.Sprintf("%04d", episode.Episode), -1)

	return filename, nil
}

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

func listEpisodes(podcastTitle string, episodes []rss.Item) {
	term.PrintLine()
	fmt.Printf("  %s\n", podcastTitle)
	term.PrintLine()

	for _, episode := range episodes {
		fmt.Printf("  Title:   %s\n", episode.Title)
		fmt.Printf("  Date:    %s\n", episode.PubDate)
		fmt.Printf("  Episode: %d\n", episode.Episode)
		fmt.Printf("  Type:    %s\n", episode.Enclosure.Type)
		term.PrintLine()
	}
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
func parseInputTime(input string) (time.Time, error) {
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
