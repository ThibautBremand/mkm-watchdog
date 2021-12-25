package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/pelletier/go-toml/v2"
	"io/ioutil"
	"text/template"
)

type Config struct {
	Delay    int
	Message  string
	Searches []Search
}

type Search struct {
	URL string
}

// Load loads the toml config, and the .env file. It returns the Config struct with the values from the toml file.
func Load() (Config, error) {
	cfg, err := loadConfig()
	if err != nil {
		return Config{}, err
	}

	err = loadEnv()
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// LoadTemplate loads the message template from the config file, used for the Telegram messages format.
func (c Config) LoadTemplate() (*template.Template, error) {
	return template.New("message").Parse(c.Message)
}

// loadConfig loads the toml file.
func loadConfig() (Config, error) {
	dat, err := ioutil.ReadFile("config.toml")
	if err != nil {
		return Config{}, fmt.Errorf("could not read config.toml: %s", err)
	}

	var cfg Config
	err = toml.Unmarshal(dat, &cfg)
	if err != nil {
		panic(err)
	}

	return cfg, nil
}

// loadEnv loads the env file.
func loadEnv() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("could not load .env: %s", err)
	}

	return nil
}
