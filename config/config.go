package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type QHConfig struct {
	Vars map[string]string
	Apps AppCatalog
}

func (config *QHConfig) Check() []error {
	var errorList []error

	for appName, app := range config.Apps {
		pathInfo, err := os.Stat(config.Resolve(app.Path))

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

func (config *QHConfig) Resolve(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	root := config.Vars["ROOT"]
	return filepath.Join(root, path)
}

func (config *QHConfig) Expand(text string) string {
	mapFunc := func(varname string) string {
		return config.Vars[varname]
	}

	return os.Expand(text, mapFunc)
}

func (config *QHConfig) ExpandWithin(text string, app AppSetup) string {
	mapFunc := func(varname string) string {
		value, found := app.Sets[varname]
		if !found {
			return config.Vars[varname]
		}
		return config.Expand(value)
	}

	return os.Expand(text, mapFunc)
}
