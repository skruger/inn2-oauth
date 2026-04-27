package oauthclient

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ClientConfig struct {
	ClientId           string   `yaml:"client_id"`
	ClientSecret       string   `yaml:"client_secret"`
	TokenURL           string   `yaml:"token_url"`
	IdentityURL        string   `yaml:"identity_url"`
	UsernameFields     []string `yaml:"username_fields"`
	OauthTokenUsername string   `yaml:"oauth_token_username"`
}

type OauthConfig struct {
	Client map[string]ClientConfig `yaml:"client"`
}

func LoadOauthConfig(filename string) (*OauthConfig, error) {
	cfgBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	cfg := &OauthConfig{}
	err = yaml.Unmarshal(cfgBytes, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (oc *OauthConfig) GetClient(clientName string) (*OauthClient, error) {
	cfg, ok := oc.Client[clientName]
	if !ok {
		return nil, fmt.Errorf("client %s not found", clientName)
	}
	if cfg.ClientId == "" || cfg.ClientSecret == "" || cfg.TokenURL == "" || cfg.IdentityURL == "" {
		fmt.Fprintf(os.Stderr, "Client %s has missing fields: client_id=%q, client_secret=%q, token_url=%q, identity_url=%q\n", clientName, cfg.ClientId, cfg.ClientSecret, cfg.TokenURL, cfg.IdentityURL)
		return nil, fmt.Errorf("client %s has missing fields", clientName)
	}
	return NewOauthClient(cfg), nil
}
