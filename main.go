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
		fmt.Fprintf(os.Stderr, "Failed to parse app list: %s\n", err)
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

func IsConfigValid() bool {
	valid := true

	for appName, appPath := range KnownApps {
		pathInfo, err := os.Stat(appPath)

		if err != nil {
			valid = false

			fmt.Fprintf(os.Stderr, "Error on app '%s' (%s): %s\n", appName, appPath, err)
			continue
		}

		if !pathInfo.IsDir() {
			valid = false
			fmt.Fprintf(os.Stderr, "Erro on app '%s': path must be a directory, got %s\n", appName, appPath)
		}
	}

	return valid
}

func addToPath(entries []string) {
	script := "export PATH=${PATH}"

        separator := string(os.PathListSeparator)
	script += separator + strings.Join(entries, separator)

	fmt.Println(script)
}

func readConfigFiles() {
	appListJson, err := os.ReadFile("./apps.json")

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read app list: %s\n", err)
	} else {
		KnownApps = parseAppList(appListJson)
	}
}

func main() {
	args := os.Args

	if len(args) <= 1 {
		return
	}

        readConfigFiles()

	switch args[1] {
	case "check":
		IsConfigValid()
		return

	case "add":
		addToPath(loadPaths(args[2:]))

	default:
		addToPath(args[1:])
	}
    
}
