package main

import (
	"time"

	mqtt "github.com/davidfantasy/embedded-mqtt-broker"
	"github.com/davidfantasy/embedded-mqtt-broker/config"
	"github.com/davidfantasy/embedded-mqtt-broker/security"
)

type CustomAuthManager struct {
}

func main() {
	config := config.NewDefaultConfig()
	config.SessionExpiryInterval = time.Second * 10
	broker := mqtt.NewMqttServer(config)
	//添加权限管理器
	broker.SetAuthProvider(&CustomAuthManager{})
	broker.Startup()
}

func (manager *CustomAuthManager) Authenticate(username, password string) *security.Authentication {
	if username == "admin" && password == "psw" {
		return security.NewAuthentication([]security.Acl{{Topic: "admin/#", Access: security.CanSubPub}})
	} else if username == "user" && password == "psw" {
		return security.NewAuthentication([]security.Acl{{Topic: "user/#", Access: security.CanSubPub}})
	}
	return nil
}
