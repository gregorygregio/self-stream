package appconfigs

import "errors"

type config struct {
	configName, value, dataType string
}

var configs = make(map[string]config)

const (
	RootIngestFolder   = "ROOT_INGESTS_FOLDER"
	RootPackagesFolder = "ROOT_PKGS_FOLDER"
)

func LoadConfigs() {
	configs[RootIngestFolder] = config{configName: RootIngestFolder, value: "./resources/raw/", dataType: "STRING"}
	configs[RootPackagesFolder] = config{configName: RootPackagesFolder, value: "./resources/hls/", dataType: "STRING"}
}

func GetConfig(configName string) (string, error) {
	c, ok := configs[configName]
	if !ok {
		return "", errors.New("config not found")
	}

	return c.value, nil
}
