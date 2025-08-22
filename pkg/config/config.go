package config

import (
	"fmt"
	"log/slog"
	"os"

	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	OsduClient OsduClient `yaml:"osdu"`
}

type OsduClient struct {
	Provider     string `yaml:"provider"`
	AuthSettings `yaml:"auth"`
	OsduSettings `yaml:"client"`
}

type AuthSettings struct {
	ClientId        string   `yaml:"clientId"`
	ClientSecret    string   `yaml:"clientSecret"`
	TenantId        string   `yaml:"tenantId"` // Added for Azure authentication
	Scopes          []string `yaml:"scopes"`
	TokenUrl        string   `yaml:"tokenUrl"`
	RefreshToken    string   `yaml:"refreshToken"`
	GrantType       string   `yaml:"grantType"`
	InternalService bool     `yaml:"internal"`
	SdkAuth         bool     `yaml:"sdkAuth"` // Added for Azure SDK Authentication (Managed Identity, etc.)
}

type OsduSettings struct {
	DatasetUrl         string `yaml:"datasetUrl"`
	PartitionUrl       string `yaml:"partitionUrl"`
	EntitlementsUrl    string `yaml:"entitlementsUrl"`
	WorkflowUrl        string `yaml:"workflowUrl"`
	SchemaUrl          string `yaml:"schemaUrl"`
	EntitlementsDomain string `yaml:"entitlementsDomain"`
	PartitionId        string `yaml:"partitionId"`
	PartitionOverrides string `yaml:"partitionOverrides"`
}

func GetAuthSettings() (AuthSettings, error) {
	config_file := GetConfigFile()
	cfg, err := GetConfig(config_file)
	if err != nil {
		slog.Error(err.Error())
		return AuthSettings{}, err
	}

	cfg.OsduClient.AuthSettings.ClientId = SetEnvSetting("OSDU_AUTH_CLIENT_ID",
		cfg.OsduClient.AuthSettings.ClientId)
	cfg.OsduClient.AuthSettings.ClientSecret = SetEnvSetting("OSDU_AUTH_CLIENT_SECRET",
		cfg.OsduClient.AuthSettings.ClientSecret)

	return cfg.OsduClient.AuthSettings, nil
}

func GetOsduSettings() (OsduSettings, error) {
	config_file := GetConfigFile()
	cfg, err := GetConfig(config_file)
	if err != nil {
		slog.Error(err.Error())
		return OsduSettings{}, err
	}
	return cfg.OsduClient.OsduSettings, nil
}

func SetEnvSetting(envVar string, def string) string {
	if val, ok := os.LookupEnv(envVar); ok {
		return val
	}
	return def
}

func GetConfig(config_file string) (Config, error) {
	yamlFile, err := os.ReadFile(config_file)
	if err != nil {
		slog.Error(err.Error())
		return Config{}, err
	}

	var c Config
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		slog.Error(err.Error())
		return Config{}, err
	}
	return c, nil
}

func GetConfigFile() string {
	config_file_location := SetEnvSetting("CONFIG_FILE", "./config/default.yaml")
	slog.Info(fmt.Sprintf("Getting config from %s", config_file_location))
	return config_file_location
}
