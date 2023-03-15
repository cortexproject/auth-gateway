package gateway

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Credentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	TenantID string `yaml:"tenantID"`
}

type Config struct {
	Basic map[string]Credentials `yaml:"basic"`
}

func ReadYaml(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
