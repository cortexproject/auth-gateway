package gateway

import (
	"log"
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

var tenants = make([]Tenant, 0)
var isTenantsInitialized = false

func InitTenants(filePath string) {
	data, err := readYaml(filePath)
	if err != nil {
		panic(err)
	}

	tenantData := data.Basic
	for key, val := range tenantData {
		tenants = append(tenants, Tenant{
			ID:       val.ID,
			Username: key,
			Password: val.Password,
		})
	}
	isTenantsInitialized = true
}

func GetTenants() []Tenant {
	if !isTenantsInitialized {
		log.Println("No tenant is provided, returning an empty slice")
		return tenants
	}

	return tenants
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
