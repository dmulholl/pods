package sanitize

import (
	"regexp"
	"strings"

	"github.com/gosimple/slug"
)

var re = regexp.MustCompile(`[<>:"/\\|?*]+`)

func Filename(input string) string {
	return strings.TrimSpace(
		strings.Trim(
			re.ReplaceAllString(input, "-"),
			"-",
		),
	)
}

func Slug(input string) string {
	return slug.Make(input)
}
