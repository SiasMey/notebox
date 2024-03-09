package nbx

import (
	"fmt"
	"os"
	"regexp"
)

var tag_pattern = regexp.MustCompile(`#[a-zA-Z0-9-_]+|#\[\[[a-zA-Z0-9-_]+\]\]`)
var tag_strip_pattern = regexp.MustCompile(`#([a-zA-Z0-9-_]+)|#\[\[([a-zA-Z0-9-_]+)\]\]`)

const usage = `usage: nbx [subcommand]

subcommands:
  tags: lists all tags in notebox`

func Main() int {
	if len(os.Args[1:]) < 1 {
		fmt.Fprintln(os.Stderr, usage)
		return 1
	}

	content, err := os.ReadFile("test1.md")
	if err != nil {
		return 1
	}

	content_str := string(content)
	for _, tag := range getTagsFromString(content_str) {
		fmt.Fprintln(os.Stdout, tag)
	}
	return 0
}

func getTagsFromString(content_str string) []string {
	return stripTags(tag_pattern.FindAllString(content_str, -1))
}

func stripTags(tags []string) []string {
	for i, tag := range tags {
		tags[i] = tag_strip_pattern.ReplaceAllString(tag, "$1$2")
	}
	return tags
}
