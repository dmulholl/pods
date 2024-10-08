# Pods

[1]: https://castos.com/tools/find-podcast-rss-feed/
[2]: https://github.com/dmulholl/pods/releases


A simple command-line utility for downloading podcast episodes.


## Example

Download all episodes of a podcast for the month of July 2024:

```
pods --download --url <podcast-url> --after "2024-07-01" --before "2024-08-01"
```

You should insert the URL for the podcast's RSS feed.
You can use a service like [castos][1] to find the appropriate URL.


## Download

You can download a pre-compiled binary from the [releases][2] page.


## Build

Pods is written in Go. If you have a Go compiler installed, you can build and install Pods by running:

```
go install github.com/dmulholl/pods/cmd/pods@latest
```

This will download, compile, and install the latest version of the application to your `$GOPATH/bin` directory.


## Usage

Run `pods --help` to view the command line help:

```
Pods v0.5.0

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

  If no time is specified, the time defaults to 00:00:00. If no timezone offset
  is specified, the timestamp is assumed to be UTC.

  The output filename can be customized using the -f/--format option. The
  following format specifiers are supported:

  - {{episode}}:  Episode number.
  - {{episode2}}: Episode number with zero-padding, min-width: 2 digits.
  - {{episode3}}: Episode number with zero-padding, min-width: 3 digits.
  - {{episode4}}: Episode number with zero-padding, min-width: 4 digits.
  - {{ext}}:      The default file extension for the file type, e.g. '.mp3'.
  - {{season}}:   Season number.
  - {{season2}}:  Season number with zero-padding, min-width: 2 digits.
  - {{season3}}:  Season number with zero-padding, min-width: 3 digits.
  - {{season4}}:  Season number with zero-padding, min-width: 4 digits.
  - {{title}}:    The episode title.

  The default filename format is '{{episode4}}. {{title}}{{ext}}'.

  Use the --debug flag to investigate problem downloads. In debug mode, the
  application won't download any episodes. Instead it will simply list all
  available metadata for the episodes which would be downloaded.

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
  -s, --season <number>     Download episodes from the specified season.
                            This option can be specified multiple times.
  -u, --url <url>           Specifies a source URL for the RSS feed.

Flags:
      --debug               Print all metadata for episodes.
  -d, --download            Download podcast episodes.
  -h, --help                Print the application's help text.
  -q, --quiet               Quiet mode. Only reports errors.
  -v, --version             Print the application's version number.
```

## License

Pods is released under the Zero-Clause BSD licence (0BSD).
