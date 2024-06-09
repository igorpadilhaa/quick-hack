package config

import "fmt"

type AppCatalog map[string]AppSetup

type AppSetup struct {
	Name string
	Path string
	Uses []string
	Sets map[string]string
}

func (catalog AppCatalog) Add(app AppSetup) {
	catalog[app.Name] = app
}

func (catalog AppCatalog) Get(appName string) (AppSetup, bool) {
	app, found := catalog[appName]
	return app, found
}

func (catalog AppCatalog) Required(appName string) ([]AppSetup, error) {
	var apps []AppSetup

	app, found := catalog.Get(appName)
	if !found {
		return nil, fmt.Errorf("app '%s' not found", appName)
	}

	for _, depName := range app.Uses {
		dep, found := catalog.Get(depName)
		if !found {
			return nil, fmt.Errorf("could not resolve %s dependency '%s'", appName, depName)
		}
		apps = append(apps, dep)
	}

	return apps, nil
}

func (catalog AppCatalog) AllRequired(appName string) ([]AppSetup, error) {
	var requiredApps []AppSetup
	
	var appsToAdd []string
	addedApps := map[string]bool{}

	appsToAdd = append(appsToAdd, appName)

	for _, depName := range appsToAdd {
		_, alreadyAdded := addedApps[depName]
		if alreadyAdded {
			continue
		}

		dep, found := catalog.Get(appName)
		if !found {
			return nil, fmt.Errorf("required app '%s' not found", depName)
		}

		requiredApps = append(requiredApps, dep)
		addedApps[depName] = true
	}

	return requiredApps, nil
}