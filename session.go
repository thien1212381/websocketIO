package socket

import (
	"gopkg.in/olahol/melody.v1"
	"strings"
)

type Session struct {
	*melody.Session
}

func newSession(session *melody.Session) *Session {
	return &Session{session}
}

func (s *Session) Emit(key string, data interface{}) {
	msgByte := prepareData(key,data)
	s.Write(msgByte)
}

func (s *Session) GetNameSpace() string {
	path := s.Request.URL.Path
	splitD := strings.Split(path, "/")
	if len(splitD) == 4 {
		return splitD[2]
	}
	return "global"
}
