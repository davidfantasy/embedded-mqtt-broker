package config

import "time"

type ServerConfig struct {
	Address               string
	Port                  int
	SessionExpiryInterval time.Duration
}

func NewDefaultConfig() *ServerConfig {
	return &ServerConfig{
		//服务监听地址
		Address: "0.0.0.0",
		//服务监听端口
		Port: 1883,
		//默认的会话超时时间，客户端断联超过该时间后，其订阅信息及其它与会话绑定的消息都将被清除
		SessionExpiryInterval: time.Hour * 2,
	}
}
