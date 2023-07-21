package main

import (
	"fmt"
	"os"
	"strings"
)

var KnownApps = map[string]string {
    "java": "path/to/jdk",
    "node": "path/to/node",
}

func loadPaths(apps []string) []string {
    var paths []string

    for _, appName := range apps {
        appPath, exists := KnownApps[appName]

        if !exists {
            fmt.Fprintf(os.Stderr, "Unknown app '%s'\n", appName)
            continue
        }

        paths = append(paths, appPath)
    }

    return paths
}

func main() {
	args := os.Args

	if len(args) <= 1  {
		return
	}

	script := "export PATH=${PATH}"
        var newEntries []string

        if args[1] == "add" {
            newEntries = loadPaths(args[2:])
        } else {
            newEntries = args[1:]
        }

        separator := string(os.PathListSeparator)
        script += separator + strings.Join(newEntries, separator)

        fmt.Println(script)
}
