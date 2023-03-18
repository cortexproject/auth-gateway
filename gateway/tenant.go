package gateway

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Tenant struct {
	ID       string `yaml:"ID"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type config struct {
	Basic map[string]Tenant `yaml:"basic"`
}

func GetTenants(filePath string) ([]Tenant, error) {
	data, err := readYaml(filePath)
	if err != nil {
		return []Tenant{}, err
	}

	tenants := make([]Tenant, 0)
	tenantData := data.Basic
	for key, val := range tenantData {
		tenants = append(tenants, Tenant{
			ID:       val.ID,
			Username: key,
			Password: val.Password,
		})
	}

	return tenants, nil
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
