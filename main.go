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

type App struct {
	Path string
	Uses []string
	Sets map[string]string
}

type AppCatalog map[string]App

type Script struct {
	content string
}

func (script *Script) Set(variable string, value string) {
	script.content += fmt.Sprintf("export %s=%s\n", variable, value)
}

func main() {
	args := os.Args

	if len(args) <= 1 {
		return
	}

	appCatalog, err := readConfigFiles()
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	script := Script{}

	switch args[1] {
	case "check":
		checkConfig(appCatalog)
		return

	case "add":
		apps, err := getApps(appCatalog, args[2:])
		if err != nil {
			log.Fatalf("ERROR: failed complete operation: %s", err)
		}

		setupApps(&script, apps)

	default:
		addToPath(&script, args[1:])
	}

	if len(script.content) != 0 {
		fmt.Println(script.content)
	}
}

func (catalog AppCatalog) ResolveDependencies(appName string) ([]App, error) {
	alreadyAdded := map[string]bool{}
	return catalog.resolveDependenciesBut(appName, alreadyAdded)
}

func (catalog AppCatalog) resolveDependenciesBut(appName string, exclude map[string]bool) ([]App, error) {
	var deps []App
	var app App
	app, exists := catalog[appName]

	if !exists {
		return nil, fmt.Errorf("unknown app %q", appName)
	}

	if exclude[appName] {
		return deps, nil
	}

	deps = append(deps, app)
	exclude[appName] = true

	for _, dependencyName := range app.Uses {
		transitiveDeps, err := catalog.resolveDependenciesBut(dependencyName, exclude)
		if err != nil {
			return nil, fmt.Errorf("in transitive dependency %q: %w", dependencyName, err)
		}

		deps = append(deps, transitiveDeps...)
	}

	return deps, nil
}

func setupApps(script *Script, apps []App) {
	paths := PathSet{}

	for _, app := range apps {
		customVars := map[string]string{
			"HQPATH": app.Path,
		}
		paths.Add(app.Path)

		for varName, value := range app.Sets {
			value = os.Expand(value, extendVarFunc(customVars))
			script.Set(varName, value)
		}
	}

	addToPath(script, paths.Entries())
}

func extendVarFunc(variables map[string]string) func(string)string {
	return func(varname string) string {
		value, exists := variables[varname]
		if !exists {
			return varname
		}
		return value
	}
}

type AppOrPath struct {
	App
	string
}

func (ap *AppOrPath) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &ap.App); err == nil {
		return nil
	}

	return json.Unmarshal(data, &ap.string)
}

func (ap *AppOrPath) ToApp() App {
	if ap.string != "" {
		ap.Path = ap.string
	}

	return ap.App
}

func currentEnvPath() PathSet {
	pathSet := PathSet{}
	path, envExist := os.LookupEnv("PATH")

	if !envExist {
		return pathSet
	}

	pathSeparator := string(os.PathListSeparator)
	pathEntries := strings.Split(path, pathSeparator)

	pathSet.AddAll(pathEntries)
	return pathSet
}

type PathSet map[string]interface{}

func (set PathSet) Add(path string) {
	set[path] = nil
}

func (set PathSet) AddAll(paths []string) {
	for _, path := range paths {
		set.Add(path)
	}
}

func (set PathSet) Entries() []string {
	var entries []string

	for entry := range set {
		entries = append(entries, entry)
	}

	return entries
}

func parseAppList(listJson []byte) (AppCatalog, error) {
	var appList map[string]AppOrPath
	err := json.Unmarshal(listJson, &appList)

	if err != nil {
		return nil, fmt.Errorf("failed to parse app list: %s", err)
	}

	appCatalog := AppCatalog{}
	for appName, appOrPath := range appList {
		appCatalog[appName] = appOrPath.ToApp()
	}

	cleanPaths(appCatalog)
	return appCatalog, nil
}

func cleanPaths(apps AppCatalog) {
	for appName, app := range apps {
		resolvedPath, err := resolvePath(app.Path)
		if err != nil {
			delete(apps, appName)
			log.Printf("ERROR: failed to resolve path from app '%s'\n", appName)
		}

		app.Path = resolvedPath
		apps[appName] = app
	}

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

func getApps(registeredApps AppCatalog, appNames []string) ([]App, error) {
	var apps []App

	for _, appName := range appNames {
		deps, err := registeredApps.ResolveDependencies(appName)
		if err != nil {
			return nil, err
		}
		apps = append(apps, deps...)
	}

	return apps, nil
}

func checkConfig(appCatalog AppCatalog) {
	for appName, app := range appCatalog {
		pathInfo, err := os.Stat(app.Path)

		if errors.Is(err, os.ErrNotExist) {
			log.Printf("WARN: path to app %q does not exist (%s)", appName, app.Path)

		} else if err != nil {
			log.Printf("ERROR: %s", err)

		} else if !pathInfo.IsDir() {
			log.Printf("WARN: path to app %q must point to a directory (%s)", appName, app.Path)
		}
	}
}

func addToPath(script *Script, entries []string) {
	newPath := currentEnvPath()
	newPath.AddAll(entries)

	if len(entries) == 0 {
		return
	}

	separator := string(os.PathListSeparator)
	path := strings.Join(newPath.Entries(), separator)

	script.Set("PATH", path)
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
