package main

import mqtt "embedded.mqtt.broker"

func main() {
	config := mqtt.NewServerOptions()
	broker := mqtt.NewMqttServer(config)
	broker.Startup()
}
