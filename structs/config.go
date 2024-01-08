package structs

import (
	"encoding/json"
	"net/url"
	"os"
)

type Config struct {
	Repositories map[string]RepoConfig `json:"repositories,omitempty"`
}

type RepoConfig struct {
	Username string `json:"username,omitempty"`
	Token    string `json:"token,omitempty"`
}

func NewConfig() *Config {
	return &Config{
		Repositories: make(map[string]RepoConfig),
	}
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var c Config
	err = json.Unmarshal(data, &c)
	return c, err
}

func (c *Config) SaveTo(dest string) {
	data, _ := json.MarshalIndent(c, "", "  ")
	_ = os.WriteFile(dest, data, 0644)
}

func (c *Config) AddRepo(repo string, config RepoConfig) {
	c.Repositories[repo] = config
}

func (c *Config) GetCredsForRepo(repo string) (string, string) {
	if conf, ok := c.Repositories[repo]; ok {
		return conf.Username, conf.Token
	}
	return "anonymous", ""
}

func (c *Config) GenerateCredsForURL(repo string) (string, error) {
	link, err := url.Parse(repo)
	if err != nil {
		return "", err
	}
	link.User = url.UserPassword(c.GetCredsForRepo(link.Host))
	return link.String(), err
}
