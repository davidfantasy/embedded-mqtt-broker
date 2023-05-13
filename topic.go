package mqtt

import (
	"strings"
	"sync"
)

const MULTI_WILDCARD = "#"

const SINGLE_WILDCARD = "+"

const TOPIC_PART_SPLITTER = "/"

type TopicTrie struct {
	isEnded  bool
	value    string
	level    int
	parent   *TopicTrie
	children map[string]*TopicTrie
	clients  map[string]interface{}
	topic    string
}

var subscribeMu sync.Mutex

//所有已订阅topic构成的前缀树，注意第一级节点为根节点，不保存实际的topic值
var subscribedTopics *TopicTrie = &TopicTrie{value: "$root"}

var clientTopicMap map[string][]*TopicTrie = make(map[string][]*TopicTrie)

func Subscribe(topic string, clientId string) {
	if len(topic) == 0 || len(clientId) == 0 {
		return
	}
	parts := strings.Split(topic, TOPIC_PART_SPLITTER)
	subscribeMu.Lock()
	defer subscribeMu.Unlock()
	//已经订阅了则不再处理
	if hasSubscribed(topic, clientId) {
		return
	}
	trie := subscribedTopics.searchPrefix(parts)
	//当前topic还没有被添加的部分
	lackLen := len(parts) - trie.level
	//说明该topic节点已经存在了，只需要修改节点信息就行了
	if lackLen == 0 {
		trie.isEnded = true
		bindTrieAndClient(clientId, topic, trie)
	} else {
		//否则添加节点
		cur := trie.insert(parts[trie.level:])
		bindTrieAndClient(clientId, topic, cur)
	}
}

//找到某个topic的所有订阅者
func GetSubscriber(topic string) []string {
	if len(topic) == 0 {
		return nil
	}
	parts := strings.Split(topic, TOPIC_PART_SPLITTER)
	subscribeMu.Lock()
	defer subscribeMu.Unlock()
	tries := subscribedTopics.searchUseWildcard(parts)
	var clients map[string]interface{} = make(map[string]interface{})
	var clientsSlice []string
	for _, t := range tries {
		for k := range t.clients {
			if _, ok := clients[k]; !ok {
				clients[k] = struct{}{}
				clientsSlice = append(clientsSlice, k)
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
	if nodes, ok := clientTopicMap[clientId]; ok {
		count := 0
		for _, node := range nodes {
			if node.topic == topic {
				count++
				delete(node.clients, clientId)
			}
		}
		if count == len(nodes) {
			delete(clientTopicMap, clientId)
		}
	}
}

func UnsubscribeAll(clientId string) {
	if len(clientId) == 0 {
		return
	}
	subscribeMu.Lock()
	defer subscribeMu.Unlock()
	if nodes, ok := clientTopicMap[clientId]; ok {
		for _, node := range nodes {
			delete(node.clients, clientId)
		}
		delete(clientTopicMap, clientId)
	}
}

func ClearNoSubscriberTopic() {
	subscribeMu.Lock()
	defer subscribeMu.Unlock()
	clearNoSubscriberTopic(subscribedTopics)
}

//递归清除没有被任何客户端订阅的空节点
func clearNoSubscriberTopic(trie *TopicTrie) {
	for _, child := range trie.children {
		clearNoSubscriberTopic(child)
	}
	if len(trie.clients) == 0 {
		trie.isEnded = false
		if len(trie.children) == 0 && trie.parent != nil {
			delete(trie.parent.children, trie.value)
		}
	}
}

//向某个节点添加topic
func (trie *TopicTrie) insert(topicParts []string) *TopicTrie {
	cur := trie
	for _, part := range topicParts {
		if cur.children == nil {
			cur.children = make(map[string]*TopicTrie)
		}
		if cur.children[part] == nil {
			cur.children[part] = &TopicTrie{parent: cur, level: cur.level + 1, value: part}
		}
		cur = cur.children[part]
	}
	cur.isEnded = true
	return cur
}

//寻找树中与某个topic最接近的前缀节点
func (trie *TopicTrie) searchPrefix(topicParts []string) *TopicTrie {
	if len(topicParts) == 0 || trie.children == nil {
		return trie
	}
	next := trie.children[topicParts[0]]
	if next != nil {
		if len(topicParts) > 1 {
			return next.searchPrefix(topicParts[1:])
		}
		return next
	}
	return trie
}

//支持输入的topicParts带通配符或者不带通配符查找，支持字典树中的已有topic也使用通配符匹配
func (trie *TopicTrie) searchUseWildcard(topicParts []string) []*TopicTrie {
	var tries []*TopicTrie
	part := topicParts[0]
	if part == SINGLE_WILDCARD {
		for _, child := range trie.children {
			if len(topicParts) > 1 {
				tries = append(tries, child.searchUseWildcard(topicParts[1:])...)
			} else {
				if child.isEnded {
					tries = append(tries, child)
				}
			}
		}
	} else if part == MULTI_WILDCARD {
		for _, child := range trie.children {
			if child.isEnded {
				tries = append(tries, child)
			}
			tries = append(tries, child.searchUseWildcard([]string{MULTI_WILDCARD})...)
		}
	} else {
		exactlyNext := trie.children[part]
		if exactlyNext != nil {
			if len(topicParts) > 1 {
				tries = append(tries, exactlyNext.searchUseWildcard(topicParts[1:])...)
			} else {
				if exactlyNext.isEnded {
					tries = append(tries, exactlyNext)
				}
			}
		}
		//单层通配符处理
		singleWildNext := trie.children[SINGLE_WILDCARD]
		if singleWildNext != nil {
			if len(topicParts) > 1 {
				tries = append(tries, singleWildNext.searchUseWildcard(topicParts[1:])...)
			} else {
				if singleWildNext.isEnded {
					tries = append(tries, singleWildNext)
				}
			}
		}
		//多层通配符处理
		multiWildNext := trie.children[MULTI_WILDCARD]
		if multiWildNext != nil {
			if multiWildNext.isEnded {
				tries = append(tries, multiWildNext)
			}
			if len(topicParts) > 1 {
				tries = append(tries, multiWildNext.searchUseWildcard(topicParts[1:])...)
			}
		}
	}
	return tries
}

func (trie *TopicTrie) countTopic() int {
	topicTotal := 0
	if trie.isEnded {
		topicTotal++
	}
	for _, child := range trie.children {
		topicTotal += child.countTopic()
	}
	return topicTotal
}

func bindTrieAndClient(clientId string, topic string, trie *TopicTrie) {
	trie.topic = topic
	if trie.clients == nil {
		trie.clients = make(map[string]interface{})
	}
	trie.clients[clientId] = struct{}{}
	subTopics := clientTopicMap[clientId]
	clientTopicMap[clientId] = append(subTopics, trie)
}

func hasSubscribed(topic string, clientId string) bool {
	subTopics := clientTopicMap[clientId]
	if subTopics == nil {
		return false
	}
	for _, trie := range subTopics {
		if trie.topic == topic {
			return true
		}
	}
	return false
}
