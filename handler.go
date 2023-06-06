package mqtt

import (
	"fmt"
	"sync"

	"github.com/davidfantasy/embedded-mqtt-broker/client"
	"github.com/davidfantasy/embedded-mqtt-broker/logger"
	"github.com/davidfantasy/embedded-mqtt-broker/packets"
)

type MessageHandler struct {
	client *client.Client
	once   sync.Once
	//用于临时存储该客户端发送的消息
	publishMsgChan chan *packets.PublishPacket
}

func NewMessageHandler(client *client.Client) *MessageHandler {
	handler := &MessageHandler{client: client}
	handler.publishMsgChan = make(chan *packets.PublishPacket, 1000)
	handler.doForward()
	return handler
}

func (handler *MessageHandler) close() {
	handler.once.Do(func() {
		close(handler.publishMsgChan)
	})
}

//异步将客户端发送的消息转发到其它订阅者
func (handler *MessageHandler) doForward() {
	go func() {
		for packet := range handler.publishMsgChan {
			//TODO 性能优化
			subscribers := GetSubscriber(packet.TopicName)
			clients := client.FindClientsBySessionIds(subscribers)
			for _, client := range clients {
				if client != nil {
					err := packet.Write(client.Conn)
					if err != nil {
						logger.WARN.Printf("投递消息时发生错误：clientId [%s],error: %s", client.Id, err)
					}
				}
			}
		}
	}()
}

func (handler *MessageHandler) handlePublish(packet *packets.PublishPacket) error {
	if !handler.client.CanPub(packet.TopicName) {
		return nil
	}
	select {
	case handler.publishMsgChan <- packet:
	default:
		logger.WARN.Printf("数据发送频率过高，该条数据将被丢弃：%s\n", packet.String())
	}
	return nil
}

func (handler *MessageHandler) HandleMessage() error {
	for {
		packet, err := packets.ReadPacket(handler.client.Conn)
		if err != nil {
			return fmt.Errorf("read packet got error:%v", err)
		}
		if packet == nil {
			return fmt.Errorf("received nil packet")
		}
		logger.DEBUG.Printf("received packet：%s", packet.String())
		//任何消息都会刷新客户端的keepalive
		handler.client.Touch()
		switch p := packet.(type) {
		case *packets.PublishPacket:
			if err := handler.handlePublish(p); err != nil {
				return err
			}
		case *packets.SubscribePacket:
			if err := handler.handleSubscribe(p); err != nil {
				return err
			}
		case *packets.UnsubscribePacket:
			if err := handler.handleUnSubscribe(p); err != nil {
				return err
			}
		case *packets.PingreqPacket:
			if err := handler.handlePing(p); err != nil {
				return err
			}
		case *packets.DisconnectPacket:
			err := handler.handleDisconnect(p)
			return err
		default:
			return fmt.Errorf("received inappropriate packet:%v", p)
		}
	}
}

func (handler *MessageHandler) handleSubscribe(packet *packets.SubscribePacket) error {
	suback := packets.NewMqttPacket(packets.Suback).(*packets.SubackPacket)
	suback.MessageID = packet.MessageID
	suback.ReturnCodes = make([]byte, len(packet.Topics))
	for i, topic := range packet.Topics {
		if handler.client.CanSub(topic) {
			Subscribe(topic, handler.client.SessionId)
			//TODO 目前仅支持qos为0的订阅
			suback.ReturnCodes[i] = 0x00
		} else {
			//无订阅权限
			suback.ReturnCodes[i] = 0x80
		}
	}
	return suback.Write(handler.client.Conn)
}

func (handler *MessageHandler) handleUnSubscribe(packet *packets.UnsubscribePacket) error {
	unsuback := packets.NewMqttPacket(packets.Unsuback).(*packets.UnsubackPacket)
	unsuback.MessageID = packet.MessageID
	for _, topic := range packet.Topics {
		Unsubscribe(topic, handler.client.Id)
	}
	return unsuback.Write(handler.client.Conn)
}

func (handler *MessageHandler) handlePing(packet *packets.PingreqPacket) error {
	pingresp := packets.NewMqttPacket(packets.Pingresp).(*packets.PingrespPacket)
	return pingresp.Write(handler.client.Conn)
}

func (handler *MessageHandler) handleDisconnect(packet *packets.DisconnectPacket) error {
	logger.INFO.Printf("received a disconnect packet,client will be disconnect:%s", handler.client.Id)
	client.CloseClient(handler.client)
	return nil
}
