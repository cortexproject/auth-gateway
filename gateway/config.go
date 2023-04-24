package gateway

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server      ServerConfig `yaml:"server"`
	Admin       ServerConfig `yaml:"admin"`
	Tenants     []Tenant     `yaml:"tenants"`
	Distributor struct {
		URL   string   `yaml:"url"`
		Paths []string `yaml:"paths"`
	} `yaml:"distributor"`
	QueryFrontend struct {
		URL   string   `yaml:"url"`
		Paths []string `yaml:"paths"`
	} `yaml:"frontend"`
}

type ServerConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
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
