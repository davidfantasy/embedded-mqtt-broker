package mqtt

import (
	"fmt"
	"net"
	"runtime/debug"
	"sync"
	"time"

	"github.com/davidfantasy/embedded-mqtt-broker/client"
	"github.com/davidfantasy/embedded-mqtt-broker/logger"
	"github.com/davidfantasy/embedded-mqtt-broker/packets"
	"github.com/davidfantasy/embedded-mqtt-broker/security"
)

var clientMap sync.Map

type ServerOptions struct {
	Address string
	Port    int
}

type MqttServer struct {
	opt                    *ServerOptions
	authenticationProvider security.AuthenticationProvider
}

func NewMqttServer(opt *ServerOptions) *MqttServer {
	return &MqttServer{opt: opt}
}

func (server *MqttServer) SetAuthProvider(authProvider security.AuthenticationProvider) {
	server.authenticationProvider = authProvider
}

func NewServerOptions() *ServerOptions {
	return &ServerOptions{Address: "127.0.0.1", Port: 1883}
}

//启动mqtt broker
func (s *MqttServer) Startup() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%v", s.opt.Address, s.opt.Port))
	if err != nil {
		logger.ERROR.Println("mqtt server start failed:", err)
		return
	}
	logger.INFO.Printf("Listening and serving mqtt on: %s:%v", s.opt.Address, s.opt.Port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.ERROR.Println("Accept client connection failed:", err)
			continue
		}
		go processNewConn(conn, s)
	}
}

func processNewConn(conn net.Conn, server *MqttServer) {
	defer func() {
		if err := recover(); err != nil {
			s := string(debug.Stack())
			logger.ERROR.Printf("connection panic:%v,%v", err, s)
		}
		conn.Close()
	}()
	//mqtt connect handshake
	client, err := acceptMqttConnect(conn, server)
	if err != nil {
		logger.ERROR.Println("mqtt connect err:", err)
		conn.Close()
	}
	clientMap.Store(client.Id, client)
	logger.DEBUG.Println("new client connected:", client.Id)
	msgHandler := NewMessageHandler(client)
	err = msgHandler.HandleMessage()
	if err != nil {
		logger.ERROR.Println("handle connection message err:", err)
	}
	//关闭消息处理器和客户端
	msgHandler.close()
	client.Disconnect()
}

func acceptMqttConnect(conn net.Conn, server *MqttServer) (*client.Client, error) {
	//设置读取超时时间，如果超时时间内还没有收到connect的包，则返回错误
	err := conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return nil, err
	}
	packet, err := packets.ReadPacket(conn)
	if err != nil {
		return nil, fmt.Errorf("read packet got error:%v", err)
	}
	if packet == nil {
		return nil, fmt.Errorf("received nil packet")
	}
	cp, ok := packet.(*packets.ConnectPacket)
	if !ok {
		return nil, fmt.Errorf("non-CONNECT first packet received:%s", packet.String())
	}
	//刷新读取超时时间
	err = conn.SetReadDeadline(time.Time{})
	if err != nil {
		return nil, err
	}
	//验证连接报文
	var returnCode byte = cp.Validate()
	var authentication *security.Authentication
	if returnCode == packets.Accepted {
		//验证用户权限
		if server.authenticationProvider != nil {
			authProvider := server.authenticationProvider
			authentication = authProvider.Authenticate(cp.Username, string(cp.Password))
			if authentication == nil {
				returnCode = packets.ErrRefusedBadUsernameOrPassword
			}
		}
	}
	cap := packets.NewMqttPacket(packets.Connack).(*packets.ConnackPacket)
	//TODO:暂不支持会话保持
	cap.SessionPresent = false
	cap.ReturnCode = returnCode
	err = cap.Write(conn)
	if err != nil {
		return nil, err
	}
	client := client.NewClient(cp, conn, authentication)
	return client, nil
}
