package content

import (
	"mime"
	"strings"
)

var (
	contentTypeFromExtension = map[string]string{
		"json": "application/json",
		"xml":  "application/xml",
		"html": "text/html",
		"htm":  "text/html",
	}
)

func getContentTypeFromExtension(file string) string {
	s := strings.Split(file, ".")
	if len(s) < 2 {
		return ""
	}
	ext := s[len(s)-1]
	if ct, ok := contentTypeFromExtension[ext]; ok {
		return ct
	}
	return mime.TypeByExtension("." + ext)
}
