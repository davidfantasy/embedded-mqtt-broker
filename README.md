# embedded-mqtt-broker
 go语言版本的mqtt轻量级服务器实现，开发者可以自由的在此基础上进行定制，实现更为复杂的业务功能。主要特性有：
- 不依赖任何其它中间件，非常轻量； 
- 可以灵活嵌入到其它任何go语言程序中；
- 完整支持MQTT 3.1.1协议；

# 使用方式

## 下载依赖
```
go get github.com/davidfantasy/embedded-mqtt-broker
```

## 代码示例
```go
import mqtt "github.com/davidfantasy/embedded-mqtt-broker"

func main() {
	config := config.NewDefaultConfig()
	broker := mqtt.NewMqttServer(config)
	broker.Startup()
}

```
## 配置项
```go
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
```
## 权限控制
现在mqtt broker可以指定接入客户端的访问控制权限，开发者可以自定义一个**security.AuthenticationProvider**，并根据接入客户端的验证信息返回不同的权限，包括对topic的publis和subcribe的权限。示例代码如下：
```go
import (
	mqtt "github.com/davidfantasy/embedded-mqtt-broker"
	"github.com/davidfantasy/embedded-mqtt-broker/security"
)

type CustomAuthManager struct {
}

func main() {
	config := mqtt.NewServerOptions()
	broker := mqtt.NewMqttServer(config)
	//添加权限管理器
	broker.SetAuthProvider(&CustomAuthManager{})
	broker.Startup()
}

//自定义权限管理器，实现AuthenticationProvider接口
func (manager *CustomAuthManager) Authenticate(username, password string) *security.Authentication {
	if username == "admin" && password == "psw" {
		return security.NewAuthentication([]security.Acl{{Topic: "admin/#", Access: security.CanSubPub}})
	} else if username == "user" && password == "psw" {
		return security.NewAuthentication([]security.Acl{{Topic: "user/#", Access: security.CanSubPub}})
	}
	return nil
}
```
## 日志
默认DEBUG,INFO,WARN,ERROR日志都是使用os.Stdout/os.Stderr进行输出，如果需要将日志记录到其它输出流，可为不同级别的logger提供一个自定义的实现，例如：
~~~go
type NOOPLogger struct{}

//实现日志接口
func (NOOPLogger) Println(v ...interface{})               {}
func (NOOPLogger) Printf(format string, v ...interface{}) {}

func changeLogger() {
	//将debug的logger替换为空输出，这样就不会输出debug日志
	logger.DEBUG = NOOPLogger{}
}
~~~
# todos
1. 支持保留消息（RETAIN）
2. 支持QOS大于0的消息的收发
3. 性能测试和优化

后续会不断完善相关功能
