package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type AppCatalog map[string]string

func parseAppList(listJson []byte) (AppCatalog, error) {
	var appList AppCatalog

	err := json.Unmarshal(listJson, &appList)

	if err != nil {
		return nil, fmt.Errorf("failed to parse app list: %s", err)
	}

	for appName, appPath := range appList {
		appPath, err := resolvePath(appPath)

		if err != nil {
			delete(appList, appName)
			fmt.Fprintf(os.Stderr, "Error: failed to resolve path from app '%s'\n", appName)
		}

		appList[appName] = appPath
	}

	return appList, nil
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

func loadPaths(registeredApps AppCatalog, appNames []string) ([]string, error) {
	var paths []string

	for _, appName := range appNames {
		appPath, exists := registeredApps[appName]
		if !exists {
			return nil, fmt.Errorf("unknown app %q", appName)
		}

		paths = append(paths, appPath)
	}

	return paths, nil
}

func checkConfig(appCatalog AppCatalog) {
	for appName, appPath := range appCatalog {
		pathInfo, err := os.Stat(appPath)

		if errors.Is(err, os.ErrNotExist) {
			log.Printf("WARN: path to app %q does not exist (%s)", appName, appPath)

		} else if err != nil {
			log.Printf("ERROR: %s", err)

		} else if !pathInfo.IsDir() {
			log.Printf("WARN: path to app %q must point to a directory (%s)", appName, appPath)
		}
	}
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

	return parseAppList(appListJson)
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
		checkConfig(apps)
		return

	case "add":
		appPaths, err := loadPaths(apps, args[2:])
		if err != nil {
			log.Fatalf("ERROR: failed complete operation: %s", err)
		}

		addToPath(appPaths)

	default:
		addToPath(args[1:])
	}

}
