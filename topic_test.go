package mqtt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchPrefix(t *testing.T) {
	root := TopicTrie{}
	root.insert([]string{"nup", "system", "a"})
	root.insert([]string{"nup", "system", "c", "#"})
	root.insert([]string{"nup", "system", "c", "+"})
	root.insert([]string{"nup", "system", "a", "b", "c"})
	root.insert([]string{"nup", "system", "a", "b", "d"})
	finded := root.searchPrefix([]string{"nup", "system", "c", "a"})
	assert.Equal(t, finded.value, "c", "value must equal")
	assert.Equal(t, finded.level, 3, "level must equal")
	finded = root.searchPrefix([]string{"nup", "system", "a", "b"})
	assert.Equal(t, finded.value, "b", "value must equal")
	assert.Equal(t, finded.level, 4, "level must equal")
}

func TestAddSubscriber(t *testing.T) {
	Subscribe("t/a/b", "c1")
	Subscribe("t/a/c", "c1")
	Subscribe("t/b/#", "c1")
	Subscribe("t/a/+", "c2")
	Subscribe("t/+/+/a", "c2")
	//测试重复订阅
	Subscribe("t/+/+/a", "c2")
	Subscribe("t/b/#", "c3")
	Subscribe("t/c/user/1", "c3")
	Subscribe("t/c/user/2", "c3")
	assert.Equal(t, 3, len(clientTopicMap["c1"]), "topic count must equal")
	assert.Equal(t, 2, len(clientTopicMap["c2"]), "topic count must equal")
	assert.Equal(t, 3, len(clientTopicMap["c3"]), "topic count must equal")
	clients := GetSubscriber("t/c/2")
	assert.Equal(t, 0, len(clients), "client must equal")
	clients = GetSubscriber("t/c/user/1")
	assert.Equal(t, []string{"c3"}, clients, "client must equal")
	clients = GetSubscriber("t/a/b")
	assert.Equal(t, []string{"c1", "c2"}, clients, "client must equal")
	clients = GetSubscriber("t/s/m/a")
	assert.Equal(t, []string{"c2"}, clients, "client must equal")
	clients = GetSubscriber("t/b/s/m/d")
	assert.Equal(t, []string{"c1", "c3"}, clients, "client must equal")
}

func TestSearchUseWildcard(t *testing.T) {
	root := TopicTrie{}
	root.insert([]string{"nup", "system", "a"})
	root.insert([]string{"nup", "system", "c", "b"})
	root.insert([]string{"nup", "system", "#"})
	root.insert([]string{"nup", "system", "c", "+"})
	root.insert([]string{"nup", "system", "a", "b", "c"})
	root.insert([]string{"nup", "system", "a", "b", "d"})
	results := root.searchUseWildcard([]string{"nup", "system", "a"})
	assert.Equal(t, 2, len(results), "total must equal")
	results = root.searchUseWildcard([]string{"nup", "system", "c", "b"})
	assert.Equal(t, 3, len(results), "total must equal")
	results = root.searchUseWildcard([]string{"nup", "system", "a", "b"})
	assert.Equal(t, 1, len(results), "total must equal")
	results = root.searchUseWildcard([]string{"nup", "user"})
	assert.Equal(t, 0, len(results), "total must equal")
	results = root.searchUseWildcard([]string{"nup", "system", "+"})
	assert.Equal(t, 2, len(results), "total must equal")
	results = root.searchUseWildcard([]string{"nup", "system", "#"})
	assert.Equal(t, 6, len(results), "total must equal")
}

func TestUnsubscribe(t *testing.T) {
	Subscribe("t/a/b", "c1")
	Subscribe("t/a/c", "c1")
	Subscribe("t/b/#", "c1")
	Subscribe("t/a/+", "c2")
	Subscribe("t/+/+/a", "c2")
	Subscribe("t/+/+/a", "c2")
	Subscribe("t/b/#", "c3")
	Subscribe("t/c/user/1", "c3")
	Subscribe("t/c/user/2", "c3")
	clients := GetSubscriber("t/a/b")
	assert.Equal(t, []string{"c1", "c2"}, clients, "client must equal")
	Unsubscribe("t/a/b", "c1")
	clients = GetSubscriber("t/a/b")
	assert.Equal(t, []string{"c2"}, clients, "client must equal")
	Unsubscribe("t/a/+", "c2")
	clients = GetSubscriber("t/a/b")
	assert.Equal(t, 0, len(clients), "client must equal")
	UnsubscribeAll("c3")
	clients = GetSubscriber("t/c/user/1")
	assert.Equal(t, 0, len(clients), "client must equal")
	clients = GetSubscriber("t/c/user/2")
	assert.Equal(t, 0, len(clients), "client must equal")
	//测试重复移除
	Unsubscribe("t/a/b", "c1")
	//测试不存在的客户端ID
	Unsubscribe("t/a/b", "not_existed_client")
	//测试移除不存在的topic
	Unsubscribe("t/s", "c1")
}

func TestClearNoSubscriberTopic(t *testing.T) {
	Subscribe("t/a/b", "c1")
	Subscribe("t/a/c", "c1")
	Subscribe("t/b/#", "c1")
	Subscribe("t/a/+", "c2")
	Subscribe("t/+/+/a", "c2")
	Subscribe("t/b/#", "c3")
	Subscribe("t/c/user/1", "c3")
	Subscribe("t/c/user/2", "c3")
	assert.Equal(t, 7, subscribedTopics.countTopic(), "total must equal")
	UnsubscribeAll("c3")
	ClearNoSubscriberTopic()
	assert.Equal(t, 5, subscribedTopics.countTopic(), "total must equal")
}
