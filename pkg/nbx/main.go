package nbx

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var tag_pattern = regexp.MustCompile(`#[a-zA-Z0-9-_]+`)

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
	for _, tag := range getTagsFromFile(content_str) {
		fmt.Fprintln(os.Stdout, strings.Trim(tag, "#"))
	}
	return 0
}

func getTagsFromFile(content_str string) []string {
	return tag_pattern.FindAllString(content_str, -1)
}
