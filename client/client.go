package client

import (
	"net"
	"sync"
	"time"

	"github.com/davidfantasy/embedded-mqtt-broker/logger"
	"github.com/davidfantasy/embedded-mqtt-broker/packets"
)

const (
	Unknown      = 0
	Connecting   = 1
	Connected    = 2
	Disconnected = 3
)

type Client struct {
	Id            string
	status        int
	statusMutex   sync.Mutex
	Conn          net.Conn
	LastPingTime  time.Time
	ConnectedTime time.Time
	Keepalive     uint16
	pingChan      chan struct{}
}

func NewClient(cp *packets.ConnectPacket, conn net.Conn) *Client {
	logger.DEBUG.Println("新客户端连接头为：%s", cp.String())
	client := &Client{Id: cp.ClientId, ConnectedTime: time.Now(), status: Connected, Conn: conn, Keepalive: cp.Keepalive}
	client.pingChan = make(chan struct{})
	if client.Keepalive != 0 {
		client.checKeepalive()
	}
	return client
}

func (client *Client) IsConnected() bool {
	client.statusMutex.Lock()
	defer client.statusMutex.Unlock()
	return client.status == Connected
}

func (client *Client) Touch() {
	if client.Keepalive != 0 {
		client.pingChan <- struct{}{}
	}
}

func (client *Client) Disconnect() {
	if !client.IsConnected() {
		return
	}
	client.statusMutex.Lock()
	defer client.statusMutex.Unlock()
	err := client.Conn.Close()
	if err != nil {
		client.status = Unknown
		logger.ERROR.Printf("close client connection err: %v \n", err)
		return
	}
	client.status = Disconnected
}

func (client *Client) checKeepalive() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.ERROR.Printf("在checKeepalive时候发生了一个错误：%v", err)
			}
		}()
		ticker := time.NewTicker(time.Duration(client.Keepalive) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if !client.IsConnected() {
					return
				}
				pingDelay := time.Since(client.LastPingTime)
				if pingDelay >= time.Duration(client.Keepalive)*time.Second*3/2 {
					logger.INFO.Printf("client：%v 在规定的周期内没有收到客户端的有效消息，准备断开连接", client.Id)
					client.Disconnect()
					return
				}
			case <-client.pingChan:
				client.LastPingTime = time.Now()
			}
		}
	}()
}
