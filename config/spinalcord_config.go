package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
)

type SpinalCordConfig struct {
	Core   CoreConfig          `json:"core"`
	Inputs map[string]IOConfig `json:"inputs"`
	//Outputs map[string]IOConfig `json:"outputs"`
	Reactor ReactorConfig `json:"reactor"`
}

func (s *SpinalCordConfig) Validate() error {
	if s.Core.HandlersDir == "" {
		return fmt.Errorf("Handler directory required! (handlers_dir)")
	}

	if _, err := os.Stat(s.Core.HandlersDir); err != nil {
		return fmt.Errorf("Could not open handlers directory: '%s'; Reason: %s\n",
			s.Core.HandlersDir, err)
	}

	if s.Core.Web.Webroot != "" {
		if _, err := os.Stat(s.Core.Web.Webroot); err != nil {
			return fmt.Errorf("Could not open webroot: '%s'; Reason: %s\n",
				s.Core.Web.Webroot, err)
		}
	}
	return nil
}

type ReactorConfig struct {
	Enabled       bool                    `toml:"enabled" json:"enabled"`
	Port          int                     `toml:"port" json:"port"`
	CreateSamples bool                    `toml:"create_samples" json:"create_samples"`
	SpinalCord    ReactorSpinalCordConfig `json:"spinalcord"`
}

type ReactorSpinalCordConfig struct {
	URI           string   `toml:"uri" json:"uri"`
	Type          string   `toml:"type" json:"type"`
	Subscriptions []string `toml:"subscriptions" json:"subscriptions"`
}

type CoreWebConfig struct {
	Port    int    `toml:"port" json:"port"`
	Webroot string `toml:"webroot" json:"webroot"`
}

type CoreConfig struct {
	HandlersDir string        `toml:"handlers_dir" json:"handlers_dir"`
	LogLevel    string        `toml:"log_level" json:"log_level"`
	Web         CoreWebConfig `json:"web"`
	Publisher   IOConfig      `json:"publisher"`
}

type IOConfig struct {
	Enabled bool                   `toml:"enabled" json:"enabled"`
	Type    string                 `toml:"type" json:"type"`
	Port    int                    `toml:"port" json:"port"`
	Config  map[string]interface{} `json:"config"`
}

func LoadConfigFromTomlFile(filepath string) (*SpinalCordConfig, error) {
	var cfg SpinalCordConfig
	d, err := ioutil.ReadFile(filepath)
	if err != nil {
		return &cfg, err
	}
	_, err = toml.Decode(string(d), &cfg)
	if err != nil {
		return &cfg, err
	}
	return &cfg, nil
}
