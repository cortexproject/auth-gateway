package gateway

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ServerAddress string   `yaml:"server-address"`
	Tenants       []Tenant `yaml:"tenants"`
	Distributor   struct {
		URL   string   `yaml:"url"`
		Paths []string `yaml:"paths"`
	} `yaml:"distributor"`
}

type Tenant struct {
	Authentication string `yaml:"authentication"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	ID             string `yaml:"id"`
}

func Init(filePath string) (Config, error) {
	configFile, err := os.ReadFile(filePath)
	if err != nil {
		return Config{}, err
	}

	config := Config{}
	err = yaml.UnmarshalStrict(configFile, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
