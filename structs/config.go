package structs

import (
	"encoding/json"
	"os"
)

type Config struct {
	Repositories []RepoConfig `json:"repositories,omitempty"`
}

type RepoConfig struct {
	Username string `json:"username,omitempty"`
	Token    string `json:"token,omitempty"`
}

func NewConfig() Config {
	return Config{
		Repositories: []RepoConfig{},
	}
}

func (c *Config) SaveTo(dest string) {
	data, _ := json.MarshalIndent(c, "", "  ")
	_ = os.WriteFile(dest, data, 0644)
}
