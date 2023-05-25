package trie

import (
	"strings"

	"github.com/davidfantasy/embedded-mqtt-broker/consts"
)

const MULTI_WILDCARD = "#"

const SINGLE_WILDCARD = "+"

//支持通配符的字典树
type TopicTrie struct {
	isEnded  bool
	topic    string
	symbol   string
	level    int
	parent   *TopicTrie
	children map[string]*TopicTrie
	Value    interface{}
	//当前节点的引用数量，每次insert加1，每次remove减去1，如果为0就删除节点
	refs int
}

func NewRootTopicTrie() *TopicTrie {
	return &TopicTrie{symbol: "$root", level: 0}
}

//向某个节点添加topic
func (trie *TopicTrie) Insert(parts []string, value interface{}) *TopicTrie {
	cur := trie
	for _, part := range parts {
		if cur.children == nil {
			cur.children = make(map[string]*TopicTrie)
		}
		if cur.children[part] == nil {
			cur.children[part] = &TopicTrie{parent: cur, level: cur.level + 1, Value: value, symbol: part}
		}
		cur = cur.children[part]
	}
	cur.isEnded = true
	cur.topic = strings.Join(parts, consts.TOPIC_PART_SPLITTER)
	cur.refs += 1
	return cur
}

//寻找字典树中层级与某个topic最接近的前缀节点
func (trie *TopicTrie) SearchPrefix(parts []string) *TopicTrie {
	if len(parts) == 0 || trie.children == nil {
		return trie
	}
	next := trie.children[parts[0]]
	if next != nil {
		if len(parts) > 1 {
			return next.SearchPrefix(parts[1:])
		}
		return next
	}
	return trie
}

//查找与某个topic相匹配的所有节点,topicParts中包含的统配符只会被当作普通字符处理
func (trie *TopicTrie) MatchMany(topicParts []string) []*TopicTrie {
	results := trie.searchTrieWithMutiMatch(topicParts)
	//去重
	if len(results) > 0 {
		// Remove duplicates by creating a map with topic values as keys
		uniqueTries := make(map[string]*TopicTrie, len(results))
		for _, t := range results {
			uniqueTries[t.topic] = t
		}
		// Convert map back to list and return
		uniqueList := make([]*TopicTrie, 0, len(uniqueTries))
		for _, t := range uniqueTries {
			uniqueList = append(uniqueList, t)
		}
		return uniqueList
	}
	return results
}

func (trie *TopicTrie) searchTrieWithMutiMatch(topicParts []string) []*TopicTrie {
	var tries []*TopicTrie
	part := topicParts[0]
	exactlyNext := trie.children[part]
	if exactlyNext != nil {
		if len(topicParts) > 1 {
			tries = append(tries, exactlyNext.searchTrieWithMutiMatch(topicParts[1:])...)
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
			tries = append(tries, singleWildNext.searchTrieWithMutiMatch(topicParts[1:])...)
		} else {
			if singleWildNext.isEnded {
				tries = append(tries, singleWildNext)
			}
		}
	}
	//多层通配符处理
	multiWildNext := trie.children[MULTI_WILDCARD]
	if multiWildNext != nil {
		tries = append(tries, multiWildNext)
	}

	return tries
}

//查找与某个topic最匹配的节点，如果没有找到，则返回nil
func (trie *TopicTrie) MatchOne(topicParts []string) *TopicTrie {
	result := trie.searchTrieWithSingleMatch(topicParts)
	size := len(topicParts)
	if result == nil && size >= 2 {
		tmp := make([]string, len(topicParts))
		for i := size - 2; i > 0; i-- {
			//如果首次搜索没有找到匹配的节点，需要将待搜索的parts的再从后向前逐层替换再进行查找
			//去匹配之前搜索中被跳过的含有通配符的节点，
			//注意：向前逐层搜索时只应匹配含通配符的节点（精确匹配只在第一次搜索时已经完成了）
			//所以只需将除了首位两位以外的其它各位分别替换为+进行匹配即可
			if result == nil {
				copy(tmp, topicParts)
				tmp[i] = SINGLE_WILDCARD
				result = trie.searchTrieWithSingleMatch(tmp)
			} else {
				break
			}
		}
	}
	return result
}

func (trie *TopicTrie) searchTrieWithSingleMatch(topicParts []string) *TopicTrie {
	part := topicParts[0]
	exactlyNext := trie.children[part]
	if exactlyNext != nil {
		if len(topicParts) > 1 {
			return exactlyNext.searchTrieWithSingleMatch(topicParts[1:])
		} else {
			if exactlyNext.isEnded {
				return exactlyNext
			}
		}
	}
	//单层通配符处理
	singleWildNext := trie.children[SINGLE_WILDCARD]
	if singleWildNext != nil {
		if len(topicParts) > 1 {
			return singleWildNext.searchTrieWithSingleMatch(topicParts[1:])
		} else {
			if singleWildNext.isEnded {
				return singleWildNext
			}
		}
	}
	//多层通配符处理
	multiWildNext := trie.children[MULTI_WILDCARD]
	if multiWildNext != nil {
		return multiWildNext
	}
	return nil
}

//从字典树中删除一个节点
func (trie *TopicTrie) Remove(parts []string) bool {
	if len(parts) == 0 {
		return false
	}
	node := trie.SearchPrefix(parts)
	if node == nil {
		return false
	}
	node.refs -= 1
	if node.refs == 0 {
		if len(node.children) == 0 {
			delete(node.parent.children, node.symbol)
		} else {
			node.isEnded = false
		}
	}
	return true
}

func (trie *TopicTrie) CountNodes() int {
	total := 0
	if trie.isEnded {
		total++
	}
	for _, child := range trie.children {
		total += child.CountNodes()
	}
	return total
}

func (trie *TopicTrie) GetLevel() int {
	return trie.level
}

func (trie *TopicTrie) IsEnded() bool {
	return trie.isEnded
}

func (trie *TopicTrie) GetTopic() string {
	return trie.topic
}
