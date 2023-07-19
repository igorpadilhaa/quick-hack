package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args

	if len(args) <= 1  {
		return
	}

	for _, entry := range args[1:] {
		fmt.Printf("Processing: %s\n", entry)
	}
}
