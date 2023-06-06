package mqtt

import (
	"strings"
	"sync"

	"github.com/davidfantasy/embedded-mqtt-broker/client"
	"github.com/davidfantasy/embedded-mqtt-broker/consts"
	"github.com/davidfantasy/embedded-mqtt-broker/event"
	"github.com/davidfantasy/embedded-mqtt-broker/trie"
)

var subscribeMu sync.Mutex

//所有已订阅topic构成的前缀树，注意第一级节点为根节点，不保存实际的topic值
var subscribedTopics *trie.TopicTrie = trie.NewRootTopicTrie()

var sessionTopicMap map[string][]string = make(map[string][]string)
var topicSessionMap map[string][]string = make(map[string][]string)

func init() {
	event.DefaultEventBus.Subscribe(event.SESSION_EXPIRIED, func(event event.Event) {
		session := event.Data.(*client.Session)
		UnsubscribeAll(session.Id)
	})
}

func Subscribe(topic string, sessionId string) {
	if len(topic) == 0 || len(sessionId) == 0 {
		return
	}
	parts := strings.Split(topic, consts.TOPIC_PART_SPLITTER)
	subscribeMu.Lock()
	defer subscribeMu.Unlock()
	//已经订阅了则不再处理
	if hasSubscribed(topic, sessionId) {
		return
	}
	subscribedTopics.Insert(parts, "")
	bindTopicAndSession(sessionId, topic)
}

//找到某个topic的所有订阅者
func GetSubscriber(topic string) []string {
	if len(topic) == 0 {
		return nil
	}
	parts := strings.Split(topic, consts.TOPIC_PART_SPLITTER)
	subscribeMu.Lock()
	defer subscribeMu.Unlock()
	tries := subscribedTopics.MatchMany(parts)
	var clients map[string]interface{} = make(map[string]interface{})
	var clientsSlice []string
	for _, t := range tries {
		subscribers := topicSessionMap[t.GetTopic()]
		for _, clientId := range subscribers {
			//对添加的clientId去重
			if _, ok := clients[clientId]; !ok {
				clients[clientId] = struct{}{}
				clientsSlice = append(clientsSlice, clientId)
			}
		}
	}
	return clientsSlice
}

func Unsubscribe(topic string, clientId string) {
	if len(topic) == 0 || len(clientId) == 0 {
		return
	}
	subscribeMu.Lock()
	defer subscribeMu.Unlock()
	hasSubscriber := unbindTopicAndSession(topic, clientId)
	//如果某个topic已经没有订阅者了，则从字典树中删除
	if !hasSubscriber {
		parts := strings.Split(topic, consts.TOPIC_PART_SPLITTER)
		subscribedTopics.Remove(parts)
	}
}

func UnsubscribeAll(sessionId string) {
	if len(sessionId) == 0 {
		return
	}
	subscribeMu.Lock()
	defer subscribeMu.Unlock()
	if topics, ok := sessionTopicMap[sessionId]; ok {
		//复制一份数据，避免循环时对topic进行了修改后会导致index错乱
		topicsCopy := make([]string, len(topics))
		copy(topicsCopy, topics)
		for _, topic := range topicsCopy {
			hasSubscriber := unbindTopicAndSession(sessionId, topic)
			if !hasSubscriber {
				parts := strings.Split(topic, consts.TOPIC_PART_SPLITTER)
				subscribedTopics.Remove(parts)
			}
		}
	}
}

func bindTopicAndSession(sessionId string, topic string) {
	topics := sessionTopicMap[sessionId]
	if topics == nil {
		topics = make([]string, 0)
	}
	sessionTopicMap[sessionId] = append(topics, topic)
	clients := topicSessionMap[topic]
	if clients == nil {
		clients = make([]string, 0)
	}
	topicSessionMap[topic] = append(clients, sessionId)
}

func unbindTopicAndSession(sessionId string, topic string) bool {
	//该topic是否还有订阅者
	var hasSubscriber bool = true
	topics := sessionTopicMap[sessionId]
	if topics != nil {
		topics = removeString(topics, topic)
		if len(topics) == 0 {
			delete(sessionTopicMap, sessionId)
		} else {
			sessionTopicMap[sessionId] = topics
		}
	}
	clients := topicSessionMap[topic]
	if clients != nil {
		clients = removeString(clients, sessionId)
		if len(clients) == 0 {
			delete(topicSessionMap, topic)
			hasSubscriber = false
		} else {
			topicSessionMap[topic] = clients
		}
	} else {
		hasSubscriber = false
	}
	return hasSubscriber
}

func hasSubscribed(topic string, clientId string) bool {
	subTopics := sessionTopicMap[clientId]
	if subTopics == nil {
		return false
	}
	for _, subscribedTopic := range subTopics {
		if subscribedTopic == topic {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) []string {
	for i, str := range slice {
		if str == s {
			copy(slice[i:], slice[i+1:])
			slice = slice[:len(slice)-1]
			break
		}
	}
	return slice
}
