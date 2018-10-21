package main

import (
	"github.com/aprosvetova/ninebot-mqtt/mqtt"
	"github.com/aprosvetova/ninebot-mqtt/scooter"
	"github.com/currantlabs/ble"
	"github.com/currantlabs/ble/linux"
	"log"
)

var mqttC *mqtt.Client
var connected = make(map[string]struct{})
var config *Config

func main() {
	var err error
	config, err = getConfig()
	if err != nil {
		log.Printf("can't load config file, DEFAULT VALUES will be used: %s\n", err)
	}

	d, err := linux.NewDevice()
	if err != nil {
		log.Fatalf("can't create BLE device: %s\n", err)
	}
	ble.SetDefaultDevice(d)

	mqttC, err = mqtt.Connect(mqtt.Options{
		Address: config.Address,
		Username: config.Username,
		Password: config.Password,
		BatteryTopic: config.BatteryTopic,
		AvailabilityTopic: config.AvailabilityTopic,
		PayloadAvailable: config.PayloadAvailable,
		PayloadNotAvailable: config.PayloadNotAvailable,
	})
	if err != nil {
		log.Fatalf("can't connect to MQTT: %s\n", err)
	}
	log.Println("connected to MQTT")

	log.Println("started listening")
	for {
		scooters, err := scooter.FindScooters()
		if err.Error() == "context canceled" {
			log.Fatalln("quitting...")
		}
		for _, mac := range scooters {
			if !containsKey(connected, mac) && !containsKey(config.ignoredScootersMap, mac) {
				go connect(mac)
			}
		}
	}
}

func connect(mac string) {
	connected[mac] = struct{}{}
	name := scooter.Listen(mac, handleBattery)
	if config.ScooterNaming == "mac" {
		name = mac
	}
	mqttC.SendOffline(name)
	delete(connected, mac)
}

func handleBattery(name string, mac string, percent int) {
	if config.ScooterNaming == "mac" {
		name = mac
	}
	mqttC.SendBatteryStatus(name, percent)
}

func containsKey(m map[string]struct{}, v string) bool {
	_, contains := m[v]
	return contains
}