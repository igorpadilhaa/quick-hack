package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	args := os.Args

	if len(args) <= 1  {
		return
	}

	script := "export PATH=${PATH}"
        
        separator := string(os.PathListSeparator)
        script += separator + strings.Join(args[1:], separator)

        fmt.Println(script)
}
