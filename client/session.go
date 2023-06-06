package client

import (
	"sync"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/davidfantasy/embedded-mqtt-broker/event"
	"github.com/davidfantasy/embedded-mqtt-broker/logger"
)

type Session struct {
	Id       string
	ClientId string
	ttl      time.Duration
	expireAt int64
}

var clientSessionMap map[string]*Session = make(map[string]*Session)

var sessionMap map[string]*Session = make(map[string]*Session)

var snowflakeNode *snowflake.Node

var sessionMu sync.Mutex

func init() {
	var err error
	snowflakeNode, err = snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
	go doSessionClearInterval()
}

func doSessionClearInterval() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		clearSessions()
	}
}

func createSession(clientId string, ttl time.Duration, resumeSession bool) (string, bool) {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	session, ok := clientSessionMap[clientId]
	if ok {
		if !resumeSession {
			doClearSession(session)
		} else {
			//复用session
			session.ttl = ttl
			session.expireAt = -1
			return session.Id, true
		}
	}
	session = &Session{Id: snowflakeNode.Generate().String(), ClientId: clientId, ttl: ttl, expireAt: -1}
	clientSessionMap[clientId] = session
	sessionMap[session.Id] = session
	return session.Id, false
}

//session对应的连接已断开，开始计算超时时间
func sessionInactive(clientId string, sessionId string) {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	session, ok := clientSessionMap[clientId]
	if ok {
		if session.Id == sessionId {
			session.expireAt = time.Now().Add(session.ttl).UnixMilli()
		} else {
			logger.WARN.Printf("sessionId与clientId不匹配：%v,%v", clientId, sessionId)
		}
	} else {
		logger.WARN.Printf("没有找到client对应的session：%v", clientId)
	}
}

func clearSession(clientId string, sessionId string) {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	session, ok := clientSessionMap[clientId]
	if ok {
		if session.Id == sessionId {
			doClearSession(session)
		} else {
			logger.WARN.Printf("sessionId与clientId不匹配：%v,%v", clientId, sessionId)
		}
	} else {
		logger.WARN.Printf("没有找到client对应的session：%v", clientId)
	}
}

func doClearSession(session *Session) {
	delete(clientSessionMap, session.ClientId)
	event.DefaultEventBus.Publish(event.NewEvent(event.SESSION_EXPIRIED, session))
	logger.DEBUG.Printf("session已过期清除：%v,%v", session.ClientId, session.Id)
}

func clearSessions() {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	for _, session := range clientSessionMap {
		if session.expireAt == -1 {
			continue
		}
		if time.Now().UnixMilli() > session.expireAt {
			delete(clientSessionMap, session.ClientId)
			event.DefaultEventBus.Publish(event.NewEvent(event.SESSION_EXPIRIED, session))
			logger.DEBUG.Printf("session已过期清除：%v,%v", session.ClientId, session.Id)
		}
	}
}

func findSessions(sessionIds []string) []*Session {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	if len(sessionIds) == 0 {
		return nil
	}
	sessions := make([]*Session, 0)
	for _, id := range sessionIds {
		session, ok := sessionMap[id]
		if ok {
			sessions = append(sessions, session)
		}
	}
	return sessions
}
