package config

import (
	"os"

	"github.com/ghanithan/challenge2016/instrumentation"
	"gopkg.in/yaml.v2"
)

func GetConfig(logger instrumentation.GoLogger, args ...string) (*Config, error) {
	//init config struct
	config := &Config{}

	// set default file path
	filePath := "../setting/sample.yaml"
	// collect the filepath from varidac arguments if provided
	if len(args) > 0 {
		filePath = args[0]
	}
	logger.Info(filePath)

	// read the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// unmarshal the yaml file into conifg struct
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}
