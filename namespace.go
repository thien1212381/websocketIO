package socket

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"errors"
)

type Namespace struct {
	l []*Session
	h HandleMessageListen
	isAuthen bool
	secret string
}
type FuncListen func(session *Session, data map[string]interface{})
type HandleMessageListen map[string]FuncListen


func newNameSpace(secret string) *Namespace {
	ns := &Namespace{
		l : []*Session{},
		h : make(HandleMessageListen),
		secret: secret,
	}

	if secret != "" {
		ns.isAuthen = true
		ns.On(AUTHENTICATE_MESSAGE, func(ses *Session, data map[string]interface{}) {
			result := struct {
				Token string `json:"token"`
			}{}

			BindData(data, &result)

			if _, err := ns.authenticateToken(result.Token); err == nil { //condition authenticate

				ses.Set(AUTHORIZED_MESSAGE, result.Token)

				ses.Emit(AUTHORIZED_MESSAGE, map[string]interface{}{
					AUTHORIZED_MESSAGE : true,
				})

				ns.l = append(ns.l, ses)

			} else {

				ses.Emit(AUTHORIZED_MESSAGE, map[string]interface{}{
					AUTHORIZED_MESSAGE : false,
					"error" : err.Error(),
				})

			}
		})
	}

	return ns
}

func (ns *Namespace) authenticateToken(token string) (*jwt.Token, error) {
	tokenCheck, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(ns.secret), nil
	})

	if err != nil {
		return nil, err
	} else {

		 if tokenCheck.Valid {
			 return tokenCheck, nil
		 }  else {
			 return nil, errors.New("Invalid token values")
		 }

	}

}

func NewNameSpace() *Namespace{
	return newNameSpace("")
}

func NewNameSpaceWithSecret(secret string) *Namespace {
	return newNameSpace(secret)
}

func (ns *Namespace) On(key string, handle FuncListen) {
	ns.h[key] = handle
}

func prepareData(key string, data interface{}) []byte {
	jsondata,_ := json.Marshal(data)
	mapdata := map[string]interface{}{}
	json.Unmarshal(jsondata, &mapdata)
	msg := Message{
		TypeMessage: key,
		Data: mapdata,
	}
	byteData, _ := json.Marshal(msg)
	return byteData
}
