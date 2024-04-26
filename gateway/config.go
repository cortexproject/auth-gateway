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
	Alertmanager  Upstream     `yaml:"alertmanager"`
	Ruler         Upstream     `yaml:"ruler"`
}

type Upstream struct {
	URL                             string        `yaml:"url"`
	Paths                           []string      `yaml:"paths"`
	DNSRefreshInterval              time.Duration `yaml:"dns_refresh_interval"`
	HTTPClientTimeout               time.Duration `yaml:"http_client_timeout"`
	HTTPClientDialerTimeout         time.Duration `yaml:"http_client_dialer_timeout"`
	HTTPClientTLSHandshakeTimeout   time.Duration `yaml:"http_client_tls_handshake_timeout"`
	HTTPClientResponseHeaderTimeout time.Duration `yaml:"http_client_response_header_timeout"`
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
	Passthrough    bool   `yaml:"passthrough"`
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
