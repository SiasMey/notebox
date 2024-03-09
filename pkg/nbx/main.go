package nbx

import (
	"fmt"
	"os"
)

const usage = `usage: nbx [subcommand]

subcommands:
  tags: lists all tags in notebox`

func Main() int {
	if len(os.Args[1:]) < 1 {
		fmt.Fprintln(os.Stderr, usage)
		return 1
	}

	fmt.Println("Hello to you,", os.Args[1])
	return 0
}
