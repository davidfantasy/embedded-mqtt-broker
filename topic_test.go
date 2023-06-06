package mqtt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	assert.Equal(t, 3, len(sessionTopicMap["c1"]), "topic count must equal")
	assert.Equal(t, 2, len(sessionTopicMap["c2"]), "topic count must equal")
	assert.Equal(t, 3, len(sessionTopicMap["c3"]), "topic count must equal")
	sessions := GetSubscriber("t/c/2")
	assert.Equal(t, 0, len(sessions), "session must equal")
	sessions = GetSubscriber("t/c/user/1")
	assert.Equal(t, []string{"c3"}, sessions, "session must equal")
	sessions = GetSubscriber("t/a/b")
	assert.Equal(t, []string{"c1", "c2"}, sessions, "session must equal")
	sessions = GetSubscriber("t/s/m/a")
	assert.Equal(t, []string{"c2"}, sessions, "session must equal")
	sessions = GetSubscriber("t/b/s/m/d")
	assert.Equal(t, []string{"c1", "c3"}, sessions, "session must equal")
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
	sessions := GetSubscriber("t/a/b")
	assert.Contains(t, sessions, "c1", "client must equal")
	assert.Contains(t, sessions, "c2", "client must equal")
	Unsubscribe("t/a/b", "c1")
	sessions = GetSubscriber("t/a/b")
	assert.Equal(t, []string{"c2"}, sessions, "client must equal")
	Unsubscribe("t/a/+", "c2")
	sessions = GetSubscriber("t/a/b")
	assert.Equal(t, len(sessions), 0, "client must equal")
	UnsubscribeAll("c3")
	sessions = GetSubscriber("t/c/user/1")
	assert.Equal(t, len(sessions), 0, "client must equal")
	sessions = GetSubscriber("t/c/user/2")
	assert.Equal(t, len(sessions), 0, "client must equal")
	//测试重复移除
	Unsubscribe("t/a/b", "c1")
	//测试不存在的客户端ID
	Unsubscribe("t/a/b", "not_existed_client")
	//测试移除不存在的topic
	Unsubscribe("t/s", "c1")
}
