package client

import (
	"net"
	"sync"
	"time"

	"github.com/davidfantasy/embedded-mqtt-broker/config"
	"github.com/davidfantasy/embedded-mqtt-broker/logger"
	"github.com/davidfantasy/embedded-mqtt-broker/packets"
	"github.com/davidfantasy/embedded-mqtt-broker/security"
)

var clientMap sync.Map

const (
	Unknown      = 0
	Connecting   = 1
	Connected    = 2
	Disconnected = 3
)

type Client struct {
	Id             string
	SessionId      string
	status         int
	statusMutex    sync.Mutex
	Conn           net.Conn
	authentication *security.Authentication
	LastPingTime   time.Time
	ConnectedTime  time.Time
	CleanSession   bool
	Keepalive      uint16
	pingChan       chan struct{}
	pubAuthCache   map[string]bool
}

func NewClient(cp *packets.ConnectPacket, conn net.Conn, authentication *security.Authentication, serverConfig *config.ServerConfig) (*Client, bool) {
	logger.DEBUG.Println("新客户端连接头为：%s", cp.String())
	client := &Client{Id: cp.ClientId, ConnectedTime: time.Now(), status: Connected, Conn: conn, Keepalive: cp.Keepalive}
	client.pingChan = make(chan struct{})
	client.authentication = authentication
	client.pubAuthCache = make(map[string]bool)
	client.CleanSession = cp.CleanSession
	sessionId, sessionPresent := createSession(client.Id, serverConfig.SessionExpiryInterval, !client.CleanSession)
	client.SessionId = sessionId
	if client.Keepalive != 0 {
		client.checKeepalive()
	}
	//TODO 旧的客户端应该被T掉
	clientMap.Store(client.Id, client)
	return client, sessionPresent
}

func CloseClient(client *Client) {
	client.close()
	clientMap.Delete(client.Id)
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

func (client *Client) CanSub(topic string) bool {
	if client.authentication == nil {
		return true
	} else {
		return client.authentication.CanSub(topic)
	}
}

//该方法会对pubAuthCache进行读写，需要确保非并发调用（使用map是为了提高性能）
func (client *Client) CanPub(topic string) bool {
	if client.authentication == nil {
		return true
	} else {
		can, ok := client.pubAuthCache[topic]
		if !ok {
			can = client.authentication.CanPub(topic)
			client.pubAuthCache[topic] = can
		}
		return can
	}
}

func (client *Client) close() {
	if !client.IsConnected() {
		return
	}
	client.statusMutex.Lock()
	defer client.statusMutex.Unlock()
	//处理会话
	if client.CleanSession {
		clearSession(client.Id, client.SessionId)
	} else {
		sessionInactive(client.Id, client.SessionId)
	}
	client.status = Disconnected
	err := client.Conn.Close()
	if err != nil {
		client.status = Unknown
		logger.ERROR.Printf("close client connection err: %v \n", err)
		return
	}
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
					CloseClient(client)
					return
				}
			case <-client.pingChan:
				client.LastPingTime = time.Now()
			}
		}
	}()
}

func FindClientsBySessionIds(sessionIds []string) []*Client {
	sessions := findSessions(sessionIds)
	if len(sessions) != 0 {
		clients := make([]*Client, len(sessions))
		for i, s := range sessions {
			c, ok := clientMap.Load(s.ClientId)
			if ok {
				clients[i] = c.(*Client)
			}
		}
		return clients
	}
	return nil
}
