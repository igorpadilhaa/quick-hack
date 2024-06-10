package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/igorpadilhaa/quick-hack/config"
	"github.com/igorpadilhaa/quick-hack/pack"
)

type PathSet map[string]interface{}

func (set PathSet) Add(paths ...string) {
	for _, path := range paths {
		set[path] = nil
	}
}

func (set PathSet) Entries() []string {
	var entries []string

	for entry := range set {
		entries = append(entries, entry)
	}

	return entries
}

func currentEnvPath() PathSet {
	pathSet := PathSet{}
	path, envExist := os.LookupEnv("PATH")

	if !envExist {
		return pathSet
	}

	pathSeparator := string(os.PathListSeparator)
	pathEntries := strings.Split(path, pathSeparator)

	pathSet.Add(pathEntries...)
	return pathSet
}

type Script struct {
	newPath  PathSet
	vars     map[string]string
	messages []string
}

func NewScript() Script {
	return Script{
		PathSet{},
		map[string]string{},
		[]string{},
	}
}

func (script *Script) Set(variable string, value string) {
	script.vars[variable] = value
}

func (script *Script) AddToPath(entry string) {
	script.newPath.Add(entry)
}

func (script *Script) Echo(message ...string) {
	script.messages = append(script.messages, strings.Join(message, " "))
}

func (script *Script) HasChanges() bool {
	return len(script.newPath.Entries()) != 0 || len(script.vars) != 0 || len(script.messages) != 0
}

func (script Script) String() string {
	if len(script.newPath) != 0 {
		path := currentEnvPath()
		path.Add(script.newPath.Entries()...)

		separator := string(os.PathListSeparator)
		pathVal := strings.Join(path.Entries(), separator)

		script.Set("PATH", pathVal)
	}

	var generated strings.Builder
	for varname, value := range script.vars {
		generated.WriteString(fmt.Sprintf("export %s=%q;\n", varname, value))
	}

	for _, message := range script.messages {
		generated.WriteString(fmt.Sprintf("echo %q;\n", message))
	}
	return generated.String()
}

func main() {
	args := os.Args
	if len(args) <= 1 {
		return
	}

	config, err := readConfiguration()
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	script := NewScript()

	switch args[1] {
	case "check":
		errs := config.Check()
		if len(errs) != 0 {
			for _, err := range errs {
				log.Println("WARN:", err)
			}
		}

	case "add":
		apps, err := config.Apps.AllRequired(args[2])
		if err != nil {
			log.Fatalf("ERROR: failed to complete operation: %s", err)
		}
		if err := setupApps(config, &script, apps); err != nil {
			log.Fatalf("ERROR: failed to complete operation: %s", err)
		}

	case "packages":
		listPackages(config, &script)

	case "install":
		if err := installPackage(config, args[2]); err != nil {
			script.Echo("ERROR:", err.Error())
		} else {
			script.Echo(args[2], "installed")
		}
	}

	if script.HasChanges() {
		fmt.Print(script)
	}
}

func readConfiguration() (*config.QHConfig, error) {
	appsConfigPath, err := resolvePath("./apps.json")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve apps configuration path: %w", err)
	}

	configStr, err := os.ReadFile(appsConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read app list: %w", err)
	}

	conf, err := config.FromJSON(configStr)
	if err != nil {
		return nil, err
	}

	var rootPath string
	if !conf.HasRoot() {
		rootPath = "."
	} else {
		rootPath = conf.GetRoot()
	}

	rootPath, err = resolvePath(rootPath)
	if err != nil {
		return nil, err
	}
	conf.SetRoot(rootPath)

	return conf, err
}

func setupApps(config *config.QHConfig, script *Script, apps []config.AppSetup) error {
	for _, app := range apps {
		appPath := config.ExpandWithin(app.Path, app)
		appPath = config.Resolve(appPath)

		script.AddToPath(appPath)

		for varName, value := range app.Sets {
			script.Set(varName, config.ExpandWithin(value, app))
		}
	}
	return nil
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

func listPackages(conf *config.QHConfig, script *Script) {
	packageNames := conf.Packages()
	if len(packageNames) == 0 {
		script.Echo("no configured packages")
		return
	}

	script.Echo("Packages:")
	for _, name := range packageNames {
		status := ""
		installed, err := conf.IsInstalled(name)
		if err == nil && installed {
			status = "[Installed]"
		}

		script.Echo("-", name, status)
	}
}

func installPackage(conf *config.QHConfig, packageName string) error {
	app, found := conf.Apps.Get(packageName)
	if !found {
		return fmt.Errorf("unknown package %q", packageName)
	}

	if app.Package == "" {
		return fmt.Errorf("the app %q has no package configured", app.Name)
	}

	return pack.Install(app.Package, conf.ExpandWithin(app.Path, app))
}