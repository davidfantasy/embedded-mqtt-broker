# embedded-mqtt-broker
 go语言版本的mqtt轻量级服务器实现，支持MQTT 3.1.1协议，可嵌入到其它业务代码中。

# 使用方式

## 下载依赖
```
go get github.com/davidfantasy/embedded-mqtt-broker
```
## 代码示例
```go
import mqtt "github.com/davidfantasy/embedded-mqtt-broker"

func main() {
	//构建配置信息
	config := mqtt.NewServerOptions()
	//创建broker并启动
	broker := mqtt.NewMqttServer(config)
	broker.Startup()
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
# 限制&未实现的特性
1. 仅支持QOS为0的消息的收发
2. 暂不支持保留消息（RETAIN）
3. 暂不支持client的权限校验

后续会不断完善相关功能
