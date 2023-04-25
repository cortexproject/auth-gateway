package gateway

import (
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server        ServerConfig `yaml:"server"`
	Admin         ServerConfig `yaml:"admin"`
	Tenants       []Tenant     `yaml:"tenants"`
	Distributor   Upstream     `yaml:"distributor"`
	QueryFrontend Upstream     `yaml:"frontend"`
}

type Upstream struct {
	URL          string        `yaml:"url"`
	Paths        []string      `yaml:"paths"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

type ServerConfig struct {
	Address      string        `yaml:"address"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"http_server_read_timeout"`
	WriteTimeout time.Duration `yaml:"http_server_write_timeout"`
	IdleTimeout  time.Duration `yaml:"http_server_idle_timeout"`
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
