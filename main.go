package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var KnownApps map[string]string

func parseAppList(listJson []byte) map[string]string {
	var appList map[string]string

	err := json.Unmarshal(listJson, &appList)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse app list: %s\n", err)
	}

        for appName, appPath := range appList {
            appPath, err := resolvePath(appPath)

            if err != nil {
                delete(appList, appName)
                fmt.Fprintf(os.Stderr, "Error: failed to resolve path from app '%s'\n", appName)
            }

            appList[appName] = appPath
        }

	return appList
}

func resolvePath(path string) (string, error) {
    rootPath, err := os.Executable()

    if err != nil {
        return "", err
    }
    
    if filepath.IsLocal(path) {
        path = filepath.Join(rootPath, path)
    }

    path, err = filepath.Abs(path)
    return path, err
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
        if len(entries) == 0 {
            return
        }

        script := "export PATH=${PATH}"

        separator := string(os.PathListSeparator)
	script += separator + strings.Join(entries, separator)

	fmt.Println(script)
}

func readConfigFiles() {
	appListJson, err := os.ReadFile("./apps.json")

	if err != nil {
            if !errors.Is(err, os.ErrNotExist) {
	    	fmt.Fprintf(os.Stderr, "Failed to read app list: %s\n", err)
            }

            return
	}
	
        KnownApps = parseAppList(appListJson)
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
