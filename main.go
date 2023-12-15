package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type AppCatalog map[string]string

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
	rootPath = filepath.Dir(rootPath)

	if err != nil {
		return "", err
	}

	rootPath, err = filepath.Abs(rootPath)

	if err != nil {
		return "", err
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(rootPath, path)
	}

	return path, err
}

func loadPaths(registeredApps AppCatalog, appNames []string) []string {
	var paths []string

	for _, appName := range appNames {
		appPath, exists := registeredApps[appName]

		if !exists {
			fmt.Fprintf(os.Stderr, "Unknown app '%s'\n", appName)
			continue
		}

		paths = append(paths, appPath)
	}

	return paths
}

func IsConfigValid(appCatalog AppCatalog) bool {
	valid := true

	for appName, appPath := range appCatalog {
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

func readConfigFiles() (AppCatalog, error) {
	appsConfigPath, err := resolvePath("./apps.json")

	if err != nil {
		return nil, fmt.Errorf("failed to resolve apps configuration path: %w", err)
	}

	appListJson, err := os.ReadFile(appsConfigPath)

	if err != nil {
		return nil, fmt.Errorf("failed to read app list: %w", err)
	}

	return parseAppList(appListJson), nil
}

func main() {
	args := os.Args

	if len(args) <= 1 {
		return
	}
	
	apps, err := readConfigFiles()
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	switch args[1] {
	case "check":
		IsConfigValid(apps)
		return

	case "add":
		addToPath(loadPaths(apps, args[2:]))

	default:
		addToPath(args[1:])
	}

}
