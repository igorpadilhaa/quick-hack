package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type QHConfig struct {
	Vars map[string]string
	apps AppCatalog
}

func (config *QHConfig) Check() []error {
	var errorList []error

	for appName := range config.apps {
		app, err := config.App(appName)
		if err != nil {
			err = fmt.Errorf("failed to get app %s: %w", appName, err)
			errorList = append(errorList, err)
		}

		pathInfo, err := os.Stat(app.Path)

		if errors.Is(err, os.ErrNotExist) {
			err = fmt.Errorf("path to app %q does not exist (%s)", appName, app.Path)
			errorList = append(errorList, err)

		} else if err != nil {
			errorList = append(errorList, err)

		} else if !pathInfo.IsDir() {
			err := fmt.Errorf("path to app %q must point to a directory (%s)", appName, app.Path)
			errorList = append(errorList, err)
		}
	}
	return errorList
}

func (config *QHConfig) SetRoot(rootPath string) {
	config.Vars["ROOT"] = rootPath
}

func (config *QHConfig) GetRoot() string {
	return config.Vars["ROOT"]
}

func (config *QHConfig) HasRoot() bool {
	_, hasRoot := config.Vars["ROOT"]
	return hasRoot
}

func (config *QHConfig) resolve(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	root := config.Vars["ROOT"]
	return filepath.Join(root, path)
}

func (config *QHConfig) App(appName string) (AppSetup, error) {
	app, found := config.apps.Get(appName)
	if !found {
		return app, fmt.Errorf("app %q not found", appName)
	}
	app.Path = config.expandWithin(app.Path, app)
	app.Path = config.resolve(app.Path)

	for index, dir := range app.Include {
		app.Include[index] = filepath.Join(app.Path, dir)
	}

	for varname, value := range app.Sets {
		app.Sets[varname] = config.expand(value) 
	}

	if len(app.Include) == 0 {
		app.Include = append(app.Include, app.Path)
	}

	return app, nil
}

func (config *QHConfig) expand(text string) string {
	mapFunc := func(varname string) string {
		return config.Vars[varname]
	}

	return os.Expand(text, mapFunc)
}

func (config *QHConfig) expandWithin(text string, app AppSetup) string {
	mapFunc := func(varname string) string {
		value, found := app.Sets[varname]
		if !found {
			return config.Vars[varname]
		}
		return config.expand(value)
	}

	return os.Expand(text, mapFunc)
}

func (config *QHConfig) AllRequired(appName string) ([]AppSetup, error) {
	var requiredApps []AppSetup
	
	var appsToAdd []string
	addedApps := map[string]bool{}

	appsToAdd = append(appsToAdd, appName)

	for _, depName := range appsToAdd {
		_, alreadyAdded := addedApps[depName]
		if alreadyAdded {
			continue
		}

		dep, err := config.App(appName)
		if err != nil {
			return nil, err
		}

		requiredApps = append(requiredApps, dep)
		addedApps[depName] = true
	}

	return requiredApps, nil
}

func (config *QHConfig) Packages() []string {
	var packages []string

	for _, app := range config.apps {
		if app.Package != "" {
			packages = append(packages, app.Name)
		}
	}
	return packages
}

func (config *QHConfig) IsInstalled(appName string) (bool, error) {
	app, err := config.App(appName)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(app.Path)
	if err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, fmt.Errorf("failed to check package status: %w", err)
	}
}
