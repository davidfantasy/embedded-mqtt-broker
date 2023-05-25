package security

import (
	"strings"

	"github.com/davidfantasy/embedded-mqtt-broker/consts"
	"github.com/davidfantasy/embedded-mqtt-broker/logger"
	"github.com/davidfantasy/embedded-mqtt-broker/trie"
)

type AuthenticationProvider interface {
	//验证用户凭证并返回当前用户的授权信息
	Authenticate(username, password string) *Authentication
}

type Authentication struct {
	authTopicTrie trie.TopicTrie
	acls          []Acl
}

type Acl struct {
	Topic  string
	Access AccessLevel
}

type User struct {
	UserName string
	Password string
}

type AccessLevel int

const (
	CanSub AccessLevel = iota
	CanPub
	CanSubPub
)

//该权限认证器只会判断用户是否在白名单内，
//并给所有的用户统一赋予所有topic的pubsub权限
type StaticUserListAuthProvider struct {
	Users   []User
	userMap map[string]string
}

func (authProvider *StaticUserListAuthProvider) Authenticate(username, password string) *Authentication {
	p, ok := authProvider.userMap[username]
	if !ok {
		return nil
	}
	if p != password {
		return nil
	}
	return NewAuthentication([]Acl{{"#", CanSubPub}})
}

func NewStaticUserListAuthProvider(users []User) *StaticUserListAuthProvider {
	p := &StaticUserListAuthProvider{Users: users, userMap: make(map[string]string)}
	for _, user := range users {
		_, ok := p.userMap[user.UserName]
		if ok {
			logger.WARN.Printf("重复添加用户：%s\n", user.UserName)
		}
		p.userMap[user.UserName] = user.Password
	}
	return p
}

func NewAuthentication(acls []Acl) *Authentication {
	auth := Authentication{authTopicTrie: trie.TopicTrie{}, acls: acls}
	for _, acl := range acls {
		if len(acl.Topic) != 0 {
			parts := strings.Split(acl.Topic, consts.TOPIC_PART_SPLITTER)
			auth.authTopicTrie.Insert(parts, acl.Access)
		}
	}
	return &auth
}

func (auth *Authentication) CanSub(topic string) bool {
	if len(topic) == 0 {
		return false
	}
	if auth.acls == nil {
		return false
	}
	parts := strings.Split(topic, consts.TOPIC_PART_SPLITTER)
	authNode := auth.authTopicTrie.MatchOne(parts)
	if authNode == nil {
		return false
	}
	access := authNode.Value.(AccessLevel)
	return access == CanSub || access == CanSubPub
}

func (auth *Authentication) CanPub(topic string) bool {
	if len(topic) == 0 {
		return false
	}
	if auth.acls == nil {
		return false
	}
	parts := strings.Split(topic, consts.TOPIC_PART_SPLITTER)
	authNode := auth.authTopicTrie.MatchOne(parts)
	if authNode == nil {
		return false
	}
	access := authNode.Value.(AccessLevel)
	return access == CanPub || access == CanSubPub
}
