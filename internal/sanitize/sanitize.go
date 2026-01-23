package sanitize

import (
	"regexp"
	"strings"

	"github.com/gosimple/slug"
)

var _illegalCharacters = regexp.MustCompile(`[\n\t\\/<>:"|?*]+`)

func Filename(input string) string {
	output := _illegalCharacters.ReplaceAllString(input, "-")
	output = strings.Trim(output, "- ")
	return output
}

func Slug(input string) string {
	return slug.Make(input)
}
