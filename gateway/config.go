package gateway

import (
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server      Server   `yaml:"server"`
	Tenants     []Tenant `yaml:"tenants"`
	Distributor struct {
		URL          string        `yaml:"url"`
		Paths        []string      `yaml:"paths"`
		ReadTimeout  time.Duration `yaml:"read_timeout"`
		WriteTimeout time.Duration `yaml:"write_timeout"`
		IdleTimeout  time.Duration `yaml:"idle_timeout"`
	} `yaml:"distributor"`
	QueryFrontend struct {
		URL   string   `yaml:"url"`
		Paths []string `yaml:"paths"`
	} `yaml:"frontend"`
}

type Server struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

type Tenant struct {
	Authentication string `yaml:"authentication"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	ID             string `yaml:"id"`
}

type Timeouts struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
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
