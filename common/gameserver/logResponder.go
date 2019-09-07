package gameserver

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type LogResponder struct{ Responder }

var _ Responder = LogResponder{}

func NewLogResponder(r Responder) LogResponder {
	return LogResponder{r}
}

func (l LogResponder) PlayerConnected(r *http.Request) {
	log.Printf(" [%v] connected\n", r.RemoteAddr)
	l.Responder.PlayerConnected(r)
}
func (l LogResponder) PlayerUpgradeFail(r *http.Request, err error) {
	log.Printf("*[%v] upgrade failed: %v\n", r.RemoteAddr, err)
	l.Responder.PlayerUpgradeFail(r, err)
}

type Counter interface{ Count() uint }

type LogNamer interface {
	LogNameEnter() string
	LogNameLeave() string
}

type LogCountResponder struct {
	LogResponder
	counter Counter
}

var _ Responder = LogCountResponder{}

func NewLogCountResponder(r Responder, counter Counter) LogCountResponder {
	return LogCountResponder{
		NewLogResponder(r),
		counter,
	}
}

func (l LogCountResponder) PlayerJoined(c *websocket.Conn, player *BinaryPlayer) {
	l.LogResponder.PlayerJoined(c, player)
	log.Printf("+[%v] %v (%v now)\n", c.RemoteAddr(), player.Data.(LogNamer).LogNameEnter(), l.counter.Count())
}

func (l LogCountResponder) PlayerLeft(c *websocket.Conn, player *BinaryPlayer) {
	l.LogResponder.PlayerLeft(c, player)
	log.Printf("-[%v] %v (%v now)\n", c.RemoteAddr(), player.Data.(LogNamer).LogNameLeave(), l.counter.Count())
}
