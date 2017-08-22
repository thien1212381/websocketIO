package socket

import (
	"gopkg.in/olahol/melody.v1"
	"github.com/gin-gonic/gin"
	"log"
	"errors"
	"fmt"
)

const (
	AUTHENTICATE_MESSAGE 	= "authenticate"
	AUTHORIZED_MESSAGE 		= "authorized"
	CONNECTED				= "connected"
	DISCONNECTED			= "disconnected"
)

type Socket struct {
	m *melody.Melody
	r *gin.Engine
	NS map[string]*Namespace
}

func New(router *gin.Engine) *Socket{
	return &Socket{
		m : melody.New(),
		r : router,
		NS : make(map[string]*Namespace),
	}
}

func (s *Socket) Start() {

	s.r.GET("/ws", func(context *gin.Context) {
		s.m.HandleRequest(context.Writer, context.Request)
	})

	s.r.GET("/ns/:name/ws", func(context *gin.Context) {
		s.m.HandleRequest(context.Writer, context.Request)
	})

	s.NS["global"] = NewNameSpace()

	funcListenHandler := func(ses *melody.Session, byteData []byte) {
		session := newSession(ses)
		ns := session.GetNameSpace()
		namespace := s.NS[ns]
		if namespace != nil {

			msg,err := NewMessage(byteData)

			if err == nil {
				if namespace.isAuthen && msg.TypeMessage != AUTHENTICATE_MESSAGE{
					if _, isAuthorized := session.Get(AUTHORIZED_MESSAGE); !isAuthorized {
						log.Println("This session need authorized first!")
						return
					}
				}

				handle, isExist := namespace.h[msg.TypeMessage]
				if isExist {

					handle(session, msg.Data)

				} else {

					log.Println("Type message does not exist in this namespace! - ", msg.TypeMessage, " - ", ns)

				}

			} else {

				log.Println(err)

			}

		} else {
			log.Println("This namespace does not exist! - ", ns)
		}
	}

	s.m.HandleMessage(funcListenHandler)

	s.m.HandleMessageBinary(funcListenHandler)

	s.m.HandleConnect(func(ses *melody.Session) {
		session := newSession(ses)
		namespace := session.GetNameSpace()

		if s.NS[namespace] != nil {
			if !s.NS[namespace].isAuthen {
				s.NS[namespace].l = append(s.NS[namespace].l, session)
				handleConnected := s.NS[namespace].h[CONNECTED]
				if handleConnected != nil {
					handleConnected(session, map[string]interface{}{})
				}
			}
		}

	})

	s.m.HandleDisconnect(func(ses *melody.Session) {
		session := newSession(ses)
		ns := session.GetNameSpace()

		if s.NS[ns] != nil {
			index := indexInListSession(session, s.NS[ns].l)

			if index >= 0 {
				s.NS[ns].l = append(s.NS[ns].l[:index], s.NS[ns].l[index+1:]...)
				handleDisconnected := s.NS[ns].h[DISCONNECTED]
				if handleDisconnected != nil {
					handleDisconnected(session, map[string]interface{}{})
				}
			}
		}

	})

}

func (s *Socket) AddNameSpace(ns string) {
	s.NS[ns] = NewNameSpace()
}

func (s *Socket) AddNameSpaceWithSecret(ns string, secret string) {
	s.NS[ns] = NewNameSpaceWithSecret(secret)
}

func (s *Socket) broadcastToList(msg []byte, list []*Session) error {

	for _, session := range list {
		if err := session.Write(msg); err != nil {
			return err
		}
	}

	return nil
}

func (s *Socket) BroadcastToNs(namespace string, key string, data interface{}) error{
	msg := prepareData(key, data)
	if s.NS[namespace] != nil {

		return s.broadcastToList(msg, s.NS[namespace].l)

	} else {
		return errors.New(fmt.Sprintf("This namespace does not exist! - %s", namespace))
	}

}

func (s *Socket) BroadcastOtherInNs(namespace string, key string, data interface{}, session *Session) error{
	msg := prepareData(key, data)
	if s.NS[namespace] != nil {

		list := []*Session{}
		for _, val := range s.NS[namespace].l {
			if val.Session != session.Session {
				list = append(list, val)
			}
		}
		return s.broadcastToList(msg, list)

	} else {
		return errors.New(fmt.Sprintf("This namespace does not exist! - %s", namespace))
	}

}

func (s *Socket) On(namespace string,key string, handle FuncListen) {
	if namespace != "" {
		s.NS[namespace].On(key, handle)
	} else {
		for n,_ := range s.NS {
			s.NS[n].On(key, handle)
		}
	}
}

func indexInListSession(session *Session, list []*Session) int {
	for i:=0; i < len(list); i ++ {
		if session == list[i] {
			return i
		}
	}
	return -1
}
