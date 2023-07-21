package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

var KnownApps map[string]string

func parseAppList(listJson []byte) map[string]string {
    var appList map[string]string

    err := json.Unmarshal(listJson, &appList)

    if err != nil {
        fmt.Fprintf(os.Stderr,"Failed to parse app list: %s\n", err)
    }

    return appList
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

        appListJson, err := os.ReadFile("./apps.json")

        if err != nil {
            fmt.Fprintf(os.Stderr, "Failed to read app list: %s\n", err)
        } else {
            KnownApps = parseAppList(appListJson)
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
