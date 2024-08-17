package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/dmulholl/argo/v4"
	"github.com/dmulholl/pods/internal/http"
	"github.com/dmulholl/pods/internal/rss"
	"github.com/dmulholl/pods/internal/term"
)

const version = "v0.2.1"

var helptext = fmt.Sprintf(`
Pods %s

  A utility for downloading podcast episodes.

Usage:
  pods --url <rss-feed>

Description:
  By default, this utility simply lists the episodes which would be downloaded.
  Use the -d/--download flag to actually download the episodes.

  The --before and --after options accept a simple date or a full timestamp in
  RFC-3339 format, e.g.

    --after "2024-07-31"
    --after "2024-07-31T13:59:00+02:00"

  If no timezone offset is specified, the timestamp is assumed to be UTC. If no
  time is specified, it defaults to 00:00:00.

  The output filename can be customized using the -f/--format option. The
  following format specifiers are supported:

  - {{episode}}:  Episode number.
  - {{episode2}}: Episode number with zero-padding, min-width: 2 digits.
  - {{episode3}}: Episode number with zero-padding, min-width: 3 digits.
  - {{episode4}}: Episode number with zero-padding, min-width: 4 digits.
  - {{ext}}:      The default file extension for the file type, e.g. '.mp3'.
  - {{title}}:    The episode title.

  The default filename format is '{{episode4}}. {{title}}{{ext}}'.

Options:
  -a, --after <timestamp>   Download episodes published after this timestamp.
  -b, --before <timestamp>  Download episodes published before this timestamp.
  -e, --episode <number>    Download the specified episode number.
                            This option can be specified multiple times.
      --file <filepath>     Specifies a source file for the RSS feed.
  -f, --format <format>     Overrides the default format for output filenames.
                            Default: '{{episode4}}. {{title}}{{ext}}'.
  -o, --outdir <path>       Output directory for downloaded files.
                            Default: './<podcast-title>'.
  -u, --url <url>           Specifies a source URL for the RSS feed.

Flags:
  -d, --download            Download podcast episodes.
  -h, --help                Print the application's help text.
  -q, --quiet               Quiet mode. Only reports errors.
  -v, --version             Print the application's version number.
`, version)

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
	argparser.NewStringOption("format f", "{{episode4}}. {{title}}{{ext}}")
	argparser.NewFlag("download d")
	argparser.NewFlag("quiet q")

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
		before, err = parseInputTimestamp(args.StringValue("before"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to parse 'before' date: '%s'\n", args.StringValue("before"))
			return 1
		}
	}

	if args.Found("after") {
		after, err = parseInputTimestamp(args.StringValue("after"))
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

		if args.Found("quiet") {
			numErrors := downloadEpisodesQuietly(ctx, dstDirectory, args.StringValue("format"), episodes)
			if numErrors > 0 {
				return 1
			}
		} else {
			numErrors := downloadEpisodes(ctx, dstDirectory, args.StringValue("format"), episodes)
			if numErrors > 0 {
				return 1
			}
		}
	}

	return 0
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
