package scooter

import (
	"bytes"
	"context"
	"errors"
	"github.com/aprosvetova/ninebot-mqtt/scooter/protocol"
	"github.com/currantlabs/ble"
	"log"
	"strings"
	"sync"
	"time"
)

var ninebotManufacturerData = []byte{0x4e, 0x42, 0x21, 0x00, 0x00, 0x00, 0x00, 0xde}
var mutex = &sync.Mutex{}

type Scooter struct {
	mac string
	name string
	handler BatteryHandler
}

type BatteryHandler func(name string, mac string, percent int)

func FindScooters() (found []string, err error) {
	var m = make(map[string]struct{})
	mutex.Lock()
	err = ble.Scan(ble.WithSigHandler(context.WithTimeout(context.Background(), time.Second*5)), false, func(a ble.Advertisement) {
		m[a.Address().String()] = struct{}{}
	}, func(a ble.Advertisement) bool {
		data := a.ManufacturerData()
		if len(data) != 8 {
			return false
		}
		return bytes.Compare(data, ninebotManufacturerData) == 0
	})
	mutex.Unlock()
	for a := range m {
		found = append(found, a)
	}
	return
}

func Listen(mac string, handler BatteryHandler) string {
	scooter := &Scooter{mac: mac, handler: handler}
	scooter.tryConnect()
	return scooter.name
}

func (s *Scooter) tryConnect() {
	mutex.Lock()
	cln, err := ble.Connect(ble.WithSigHandler(context.WithTimeout(context.Background(), time.Second*5)), func(a ble.Advertisement) bool {
		if a.Address().String() == s.mac {
			s.name = strings.Trim(a.LocalName(), "\x00")
			return true
		}
		return false
	})
	if err != nil {
		log.Println(err)
		return
	}
	if s.name == "" {
		s.name = s.mac
	}
	s.log("connected to device")
	p, err := cln.DiscoverProfile(true)
	if err != nil {
		s.log("can't discover profile")
		return
	}
	err = s.subscribePower(cln, p)
	if err != nil {
		s.log("can't subscribe: " + err.Error())
		return
	}
	mutex.Unlock()

	for {
		select {
		case <-cln.Disconnected():
			s.log("device disconnected")
			return
		default:
			err = requestPower(cln, p)
			if err != nil {
				s.log("can't request: " + err.Error())
				return
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func requestPower(cln ble.Client, p *ble.Profile) error {
	uuid, _ := ble.Parse("6e400002b5a3f393e0a9e50e24dcca9e")
	u := p.Find(ble.NewCharacteristic(uuid))
	if u == nil {
		return errors.New("can't find characteristic")
	}
	err := cln.WriteCharacteristic(u.(*ble.Characteristic), protocol.GetBattery(), true)
	return err
}

func (s *Scooter) subscribePower(cln ble.Client, p *ble.Profile) error {
	uuid, _ := ble.Parse("6e400003b5a3f393e0a9e50e24dcca9e")
	u := p.Find(ble.NewCharacteristic(uuid))
	if u == nil {
		return errors.New("can't find characteristic")
	}
	err := cln.Subscribe(u.(*ble.Characteristic), false, func(req []byte) {
		percent := int(req[7])
		if percent < 0 || percent > 100 {
			log.Println("strange value: ", percent)
			return
		}
		s.handler(s.name, s.mac, percent)
	})
	return err
}

func (s *Scooter) log(message string) {
	log.Printf("[%s|%s] %s\n", s.name, s.mac, message)
}