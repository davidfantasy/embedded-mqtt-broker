package main

import mqtt "github.com/davidfantasy/embedded-mqtt-broker"

func main() {
	config := mqtt.NewServerOptions()
	broker := mqtt.NewMqttServer(config)
	broker.Startup()
}
