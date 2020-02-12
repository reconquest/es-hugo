package main

import "github.com/kovetskiy/ko"

type Config struct {
	Listen    string `yaml:"listen" required:"true" env:"LISTEN" default:":80"`
	Elastic   string `yaml:"elastic" required:"true" env:"ELASTIC" default:""`
	Input     string `yaml:"input" env:"INPUT" required:"true"`
	Language  string `yaml:"language" env:"LANGUAGE" required:"true"`
	Delimiter string `yaml:"delimiter" env:"DELIMITER" required:"true"`
	Index     string `yaml:"index" env:"INDEX" required:"true"`
}

func getConfig(path string) (*Config, error) {
	config := &Config{}
	err := ko.Load(path, config, ko.RequireFile(false))
	if err != nil {
		return nil, err
	}

	return config, nil
}
