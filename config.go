package main

import (
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"strings"
)

type Config struct {
	Address string
	Username string
	Password string
	BatteryTopic string `toml:"battery_topic"`
	AvailabilityTopic string `toml:"availability_topic"`
	PayloadAvailable string `toml:"payload_available"`
	PayloadNotAvailable string `toml:"payload_not_available"`
	ScooterNaming string `toml:"scooter_naming"`
	IgnoredScooters []string `toml:"ignored_scooters"`
	ignoredScootersMap map[string]struct{} `toml:"-"`
}

func getConfig() (config *Config, err error) {
	config = &Config{}
	data, err := ioutil.ReadFile("config.toml")
	if err == nil {
		err = toml.Unmarshal(data, config)
	}
	config.ignoredScootersMap = make(map[string]struct{})
	if config.Address == "" {
		config.Address = "127.0.0.1:1883"
	}
	if config.BatteryTopic == "" {
		config.BatteryTopic = "ninebot/%s/battery"
	}
	if config.AvailabilityTopic == "" {
		config.AvailabilityTopic = "ninebot/%s/available"
	}
	if config.AvailabilityTopic == "disable" {
		config.AvailabilityTopic = ""
	}
	if config.PayloadAvailable == "" {
		config.PayloadAvailable = "online"
	}
	if config.PayloadNotAvailable == "" {
		config.PayloadNotAvailable = "offline"
	}
	if config.ScooterNaming == "" {
		config.ScooterNaming = "name"
	}
	for _, ignored := range config.IgnoredScooters {
		config.ignoredScootersMap[strings.ToLower(ignored)] = struct{}{}
	}
	return
}