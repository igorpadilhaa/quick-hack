package config

import (
	"encoding/json"
)

type jsonConfig struct {
	Vars map[string]string
	Apps map[string]appOrPath
}

func (jConfig *jsonConfig) ToConfig() *QHConfig {
	var config QHConfig
	config.Vars = jConfig.Vars
	config.apps = AppCatalog{}

	for appName, appOrPath := range jConfig.Apps {
		var app AppSetup

		if appOrPath.AppSetup == nil {
			app.Path = appOrPath.string

		} else {
			app = *appOrPath.AppSetup
		}

		app.Name = appName
		config.apps.Add(app)
	}

	return &config
}

type appOrPath struct {
	*AppSetup
	string
}

func (ap *appOrPath) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &ap.AppSetup); err == nil {
		return nil
	}

	ap.AppSetup = nil
	return json.Unmarshal(data, &ap.string)
}

func FromJSON(jsonData []byte) (*QHConfig, error) {
	var jConfig jsonConfig

	if err := json.Unmarshal(jsonData, &jConfig); err != nil {
		return nil, err
	}

	return jConfig.ToConfig(), nil
}
