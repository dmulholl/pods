package sanitize

import (
	"regexp"
	"strings"
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
