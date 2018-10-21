package mqtt

import (
	"fmt"
	proto "github.com/huin/mqtt"
	"github.com/jeffallen/mqtt"
	"net"
	"strconv"
)

type Client struct {
	options Options
	mqttC *mqtt.ClientConn
}

type Options struct {
	Address string
	Username string
	Password string
	BatteryTopic string
	AvailabilityTopic string
	PayloadAvailable string
	PayloadNotAvailable string
}

func Connect(options Options) (client *Client, err error) {
	conn, err := net.Dial("tcp", options.Address)
	if err != nil {
		return
	}
	mqttC := mqtt.NewClientConn(conn)
	err = mqttC.Connect(options.Username, options.Password)
	if err != nil {
		return
	}
	client = &Client{
		options: options,
		mqttC: mqttC,
	}
	return
}

func (c *Client) sendAvailable(mac string, available bool) {
	if c.options.AvailabilityTopic == "" {
		return
	}
	data := c.options.PayloadNotAvailable
	if available {
		data = c.options.PayloadAvailable
	}
	c.mqttC.Publish(&proto.Publish{
		TopicName: fmt.Sprintf(c.options.AvailabilityTopic, mac),
		Payload:   proto.BytesPayload([]byte(data)),
	})
}

func (c *Client) SendOffline(name string) {
	c.sendAvailable(name, false)
}

func (c *Client) SendBatteryStatus(name string, percent int) {
	c.mqttC.Publish(&proto.Publish{
		TopicName: fmt.Sprintf(c.options.BatteryTopic, name),
		Payload:   proto.BytesPayload([]byte(strconv.Itoa(percent))),
	})
	c.sendAvailable(name, true)
}
