package handler

import (
	"bufio"

	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

func alarmWs(ws map[*websocket.Conn]struct{}) websocket.Handler {
	return func(conn *websocket.Conn) {
		defer func() {
			delete(ws, conn)
			conn.Close()
			log.Infof("%s has closed.", conn.RemoteAddr())
		}()
		if _, ok := ws[conn]; !ok {
			ws[conn] = struct{}{}
		}
		log.Infof("%s has conncted.", conn.RemoteAddr())
		buf := bufio.NewReader(conn)
		for {
			_, err := buf.ReadByte()
			if err != nil {
				break
			}
		}
	}
}

func AlarmHandler(ws map[*websocket.Conn]struct{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := alarmWs(ws)
		h.ServeHTTP(c.Writer, c.Request)
	}
}
