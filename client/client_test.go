package client

import (
	"strconv"
	"testing"

	"github.com/davidfantasy/embedded-mqtt-broker/config"
	"github.com/davidfantasy/embedded-mqtt-broker/packets"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	cp := packets.NewMqttPacket(packets.Connect).(*packets.ConnectPacket)
	cp.ClientId = "testC1"
	cp.CleanSession = false
	cp.Keepalive = 0
	config := config.NewDefaultConfig()
	c1, sp := NewClient(cp, nil, nil, config)
	assert.False(t, sp)
	assert.NotNil(t, c1)
	clients := FindClientsBySessionIds([]string{c1.SessionId})
	assert.Equal(t, 1, len(clients))
	assert.Equal(t, "testC1", clients[0].Id)
}

func BenchmarkFindClientsBySessionIds(b *testing.B) {
	var sessionIds []string
	for i := 0; i < 100; i++ {
		cp := packets.NewMqttPacket(packets.Connect).(*packets.ConnectPacket)
		cp.ClientId = "testC" + strconv.Itoa(i)
		cp.CleanSession = false
		cp.Keepalive = 0
		config := config.NewDefaultConfig()
		c, _ := NewClient(cp, nil, nil, config)
		sessionIds = append(sessionIds, c.SessionId)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindClientsBySessionIds(sessionIds)
	}
}
