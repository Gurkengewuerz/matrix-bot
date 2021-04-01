package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	DBFile string `yaml:"sqlite_file"`

	Homeserver string `yaml:"homeserver"`

	Bot struct {
		Username    string `yaml:"username"`
		Displayname string `yaml:"displayname"`
		Avatar      string `yaml:"avatar"`
		Password    string `yaml:"password"`
		DeviceID    string `yaml:"device_id"`
	} `yaml:"bot"`

	Plugins []string `yaml:"plugins"`

	WebServer struct {
		ListenOn string `yaml:"listen_on"`
		Port     int    `yaml:"port"`
		BaseURL  string `yaml:"base_url"`
	} `yaml:"webserver"`

	Logger struct {
		Debug  bool `yaml:"debug"`
	} `yaml:"logger"`
}

func (config *Config) setDefaults() {
	config.Bot.Displayname = "by mc8051.de"
	config.Plugins = make([]string, 0)
	config.WebServer.Port = 7785
	config.WebServer.ListenOn = "0.0.0.0"
	config.WebServer.BaseURL = fmt.Sprintf("http://%s:%v", config.WebServer.ListenOn, config.WebServer.Port)
}

func Load(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config = &Config{}
	config.setDefaults()
	err = yaml.Unmarshal(data, config)
	return config, err
}

func (config *Config) Save(path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0600)
}
