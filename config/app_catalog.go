package config

type AppCatalog map[string]AppSetup

type AppSetup struct {
	Name string
	Path string
	Include []string
	Uses []string
	Sets map[string]string
	Package string
}

func (catalog AppCatalog) Add(app AppSetup) {
	catalog[app.Name] = app
}

func (catalog AppCatalog) Get(appName string) (AppSetup, bool) {
	app, found := catalog[appName]
	return app, found
}
