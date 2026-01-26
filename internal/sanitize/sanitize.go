package sanitize

import (
	"regexp"
	"strings"

	"github.com/gosimple/slug"
)

var _illegalCharacters = regexp.MustCompile(`[\n\r\t\\/<>:"|?*]+`)

func Filename(input string) string {
	output := _illegalCharacters.ReplaceAllString(input, "-")
	output = strings.Trim(output, "- ")
	return output
}

func Slug(input string) string {
	return slug.Make(input)
}
