package gateway

import (
	"os"

	"github.com/go-kit/log"
	"gopkg.in/yaml.v2"
)

type Configuration struct {
	ServerAddress string            `yaml:"server-address"`
	AuthType      string            `yaml:"auth-type"`
	Tenants       []Tenant          `yaml:"tenants"`
	Routes        []Route           `yaml:"routes"`
	Targets       map[string]string `yaml:"targets"`
	Logger        log.Logger
}

type Tenant struct {
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	XScopeOrgId string `yaml:"x-scope-orgid"`
}

type Route struct {
	Path   string `yaml:"path"`
	Target string `yaml:"target"`
}

func Init(filePath string, logger log.Logger) (Configuration, error) {
	configFile, err := os.ReadFile(filePath)
	if err != nil {
		return Configuration{}, err
	}

	config := Configuration{}
	err = yaml.UnmarshalStrict(configFile, &config)
	if err != nil {
		return Configuration{}, err
	}

	config.Logger = logger
	return config, nil
}
