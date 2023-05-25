package trie

import (
	"fmt"
	"strings"
	"testing"

	"github.com/davidfantasy/embedded-mqtt-broker/consts"
	"github.com/stretchr/testify/assert"
)

func TestSearchPrefix(t *testing.T) {
	root := TopicTrie{}
	root.Insert([]string{"nup", "system", "a"}, "nodeA")
	root.Insert([]string{"nup", "system", "c", "#"}, "")
	root.Insert([]string{"nup", "system", "c", "+"}, "")
	root.Insert([]string{"nup", "system", "a", "b", "c"}, "nodeC")
	root.Insert([]string{"nup", "system", "a", "b", "d"}, "nodeD")
	finded := root.SearchPrefix([]string{"nup", "system", "c", "a"})
	assert.Equal(t, "c", finded.symbol, "value must equal")
	assert.Equal(t, 3, finded.level, "level must equal")
	finded = root.SearchPrefix([]string{"nup", "system", "a", "b"})
	assert.Equal(t, "b", finded.symbol, "value must equal")
	assert.Equal(t, 4, finded.level, "level must equal")
	tests := []struct {
		input    []string
		expected string
	}{
		{[]string{"nup", "system", "a"}, "nodeA"},
		{[]string{"nup", "system", "c", "b"}, ""},
		{[]string{"nup", "system", "a", "b", "c"}, "nodeC"},
		{[]string{"nup", "system", "a", "b", "d"}, "nodeD"},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("input=%v", test.input), func(t *testing.T) {
			node := root.SearchPrefix(test.input)
			if node.Value != test.expected {
				t.Errorf("expected value %v, but got %v", test.expected, node.Value)
			}
		})
	}
}

func TestMatchMany(t *testing.T) {
	root := TopicTrie{}
	root.Insert([]string{"nup", "system", "a"}, "nodeA")
	root.Insert([]string{"nup", "system", "c", "b"}, "nodeB")
	root.Insert([]string{"nup", "system", "#"}, "node#")
	root.Insert([]string{"nup", "system", "c", "+"}, "node+")
	root.Insert([]string{"nup", "system", "a", "b", "c"}, "nodeC")
	root.Insert([]string{"nup", "system", "a", "b", "d"}, "nodeD")
	results := root.MatchMany([]string{"nup", "system", "a"})
	assert.Equal(t, 2, len(results), "total must equal")
	results = root.MatchMany([]string{"nup", "system", "c", "b"})
	assert.Equal(t, 3, len(results), "total must equal")
	results = root.MatchMany([]string{"nup", "system", "a", "b"})
	assert.Equal(t, 1, len(results), "total must equal")
	results = root.MatchMany([]string{"nup", "user"})
	assert.Equal(t, 0, len(results), "total must equal")
	results = root.MatchMany([]string{"nup", "system", "+"})
	assert.Equal(t, 1, len(results), "total must equal")
	results = root.MatchMany([]string{"nup", "system", "#"})
	assert.Equal(t, 1, len(results), "total must equal")
}

func TestMatchOne(t *testing.T) {
	root := TopicTrie{}
	root.Insert([]string{"nup", "system", "a"}, "nodeA")
	root.Insert([]string{"nup", "system", "c", "b"}, "nodeB")
	root.Insert([]string{"nup", "system", "#"}, "node#")
	root.Insert([]string{"nup", "system", "c", "+"}, "node+")
	root.Insert([]string{"nup", "system", "c", "+", "s"}, "nodeS")
	root.Insert([]string{"nup", "system", "a", "b", "c"}, "nodeC")
	root.Insert([]string{"nup", "system", "a", "b", "d"}, "nodeD")
	root.Insert([]string{"nup", "system", "s", "+", "+", "m"}, "nodeM")
	results := root.MatchOne([]string{"nup", "system", "a"})
	nodeValueEQ(t, results, "nodeA")
	results = root.MatchOne([]string{"nup", "system", "c", "b"})
	nodeValueEQ(t, results, "nodeB")
	results = root.MatchOne([]string{"nup", "system", "a", "b"})
	nodeValueEQ(t, results, "node#")
	results = root.MatchOne([]string{"nup", "user"})
	assert.Nil(t, results)
	results = root.MatchOne([]string{"nup", "system", "+"})
	nodeValueEQ(t, results, "node#")
	results = root.MatchOne([]string{"nup", "system", "#"})
	nodeValueEQ(t, results, "node#")
	results = root.MatchOne(topic2parts("nup/system/a/d/+"))
	nodeValueEQ(t, results, "node#")
	results = root.MatchOne(topic2parts("nup/system/a/+"))
	nodeValueEQ(t, results, "node#")
	results = root.MatchOne(topic2parts("nup/system/c/+"))
	nodeValueEQ(t, results, "node+")
	results = root.MatchOne(topic2parts("nup/system/c/s/s"))
	nodeValueEQ(t, results, "nodeS")
	results = root.MatchOne(topic2parts("nup/system/s/s/s"))
	nodeValueEQ(t, results, "node#")
	results = root.MatchOne(topic2parts("nup/system/s/s/s/m"))
	nodeValueEQ(t, results, "nodeM")
}

func TestRemove(t *testing.T) {
	root := TopicTrie{}
	root.Insert(topic2parts("a/b/c"), "")
	root.Insert(topic2parts("a/b/c"), "")
	root.Insert(topic2parts("a/b/d"), "")
	root.Insert(topic2parts("a/c/d"), "")
	root.Insert(topic2parts("a/c/f"), "")
	root.Insert(topic2parts("a/c/e"), "")
	assert.Equal(t, root.CountNodes(), 5)
	trie := root.SearchPrefix(topic2parts("a/b/c"))
	assert.Equal(t, trie.refs, 2)
	root.Remove(topic2parts("a/b/c"))
	assert.Equal(t, trie.refs, 1)
	root.Remove(topic2parts("a/b/c"))
	trie = root.SearchPrefix(topic2parts("a/b/c"))
	assert.Equal(t, 2, trie.level)
	root.Remove(topic2parts("a/c"))
	trie = root.SearchPrefix(topic2parts("a/c/f"))
	assert.NotNil(t, trie)
	root.Remove(topic2parts("a/c/f"))
	trie = root.SearchPrefix(topic2parts("a/c/f"))
	assert.Equal(t, 2, trie.level)
}

func topic2parts(topic string) []string {
	if len(topic) == 0 {
		return []string{}
	}
	return strings.Split(topic, consts.TOPIC_PART_SPLITTER)
}

func nodeValueEQ(t *testing.T, node *TopicTrie, expectedVal interface{}) {
	if node == nil {
		assert.Fail(t, "node is nil")
		return
	}
	assert.Equal(t, expectedVal, node.Value, "Node value must equal")
}
