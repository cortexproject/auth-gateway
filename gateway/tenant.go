package gateway

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Tenant struct {
	All map[string]tenant `yaml:"All"`
}

type tenant struct {
	ID       string `yaml:"ID"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type config struct {
	Basic map[string]tenant `yaml:"basic"`
}

func GetTenants(filePath string) (Tenant, error) {
	tenants := make(map[string]tenant)
	data, err := readYaml(filePath)
	if err != nil {
		return Tenant{}, err
	}

	tenantData := data.Basic
	for key, val := range tenantData {
		tenants[key] = tenant{
			ID:       val.ID,
			Username: key,
			Password: val.Password,
		}
	}

	return Tenant{
		All: tenants,
	}, nil
}

func readYaml(filePath string) (*config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
