package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/itrabbit/just/cli"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("[WARNING] Too short arguments!")
		return
	}
	cmd := strings.ToLower(strings.TrimSpace(os.Args[1]))
	if len(cmd) < 1 {
		fmt.Println("[WARNING] Too short command!")
		return
	}
	if err := cli.RunCmd(cmd); err != nil {
		fmt.Println("[ERROR]", err)
		return
	}
	fmt.Println("Success!")
}
