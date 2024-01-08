package structs

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
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

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return NewConfig(), err
	}
	var c Config
	err = json.Unmarshal(data, &c)
	if c.Repositories == nil {
		c.Repositories = make(map[string]RepoConfig)
	}
	return &c, err
}

func (c *Config) SaveTo(dest string) {
	data, _ := json.MarshalIndent(c, "", "  ")
	_ = os.WriteFile(dest, data, 0644)
}

func (c *Config) AddRepo(repo string, config RepoConfig) {
	link, _ := url.Parse(repo)
	c.Repositories[link.Host] = config
}

func (c *Config) GetCredsForRepo(repo string) (string, string) {
	if conf, ok := c.Repositories[repo]; ok {
		return conf.Username, conf.Token
	}
	return "anonymous", ""
}

func (c *Config) GenerateRequestURL(req *http.Request, repo string) (*http.Request, error) {
	link, err := url.Parse(repo)
	if err != nil {
		return nil, err
	}
	uname, passwd := c.GetCredsForRepo(link.Host)
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", uname, passwd))),
	)
	return req, err
}
