# embedded-mqtt-broker
 go语言版本的mqtt轻量级服务器实现，支持MQTT 3.1.1，可嵌入到其它业务代码中。

# 启动方式

```go
import mqtt "embedded.mqtt.broker"

func main() {
	config := mqtt.NewServerOptions()
	broker := mqtt.NewMqttServer(config)
	broker.Startup()
}
```
# 限制&未实现的特性
1. 仅支持QOS为0的消息的收发
2. 暂不支持保留消息（RETAIN）
3. 暂不支持client的权限校验

后续会不断完善相关功能
