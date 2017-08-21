package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/thien1212381/websocketSimple"
	"log"
)



func main() {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "index.html")
	})

	s := socket.New(router)
	s.Start()

	router.GET("/ns/:name", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "chan.html")
	})

	s.AddNameSpaceWithSecret("a","secret here")
	s.AddNameSpace("b")

	for ns,_ := range s.NS {
		s.NS[ns].On("message", func(session *socket.Session, data map[string]interface{}) {
			datajson := struct {
				Msg string `json:"msg"`
			}{}
			if err := socket.BindData(data, &datajson); err != nil {
				log.Println(err)
			} else {
				log.Println(datajson)
				datajson.Msg = "test"
			}
			ns := session.GetNameSpace()
			session.Emit("message",datajson)
			s.BroadcastOtherInNs(ns, "message", data, session)
		})
	}

	router.Run(":8888")
}
