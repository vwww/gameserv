package gameserver

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type LogResponder[P any] struct{ Responder[P] }

func NewLogResponder[P any](r Responder[P]) LogResponder[P] {
	return LogResponder[P]{r}
}

func (l LogResponder[P]) PlayerConnected(r *http.Request) {
	log.Printf(" [%v] connected\n", r.RemoteAddr)
	l.Responder.PlayerConnected(r)
}
func (l LogResponder[P]) PlayerUpgradeFail(r *http.Request, err error) {
	log.Printf("*[%v] upgrade failed: %v\n", r.RemoteAddr, err)
	l.Responder.PlayerUpgradeFail(r, err)
}

type Counter interface{ Count() uint }

type LogNamer interface {
	LogNameEnter() string
	LogNameLeave() string
}

type LogCountResponder[P LogNamer] struct {
	LogResponder[P]
	counter Counter
}

func NewLogCountResponder[P LogNamer](r Responder[P], counter Counter) LogCountResponder[P] {
	return LogCountResponder[P]{
		NewLogResponder(r),
		counter,
	}
}

func (l LogCountResponder[P]) PlayerJoined(c *websocket.Conn, player *BinaryPlayer[P]) {
	l.LogResponder.PlayerJoined(c, player)
	log.Printf("+[%v] %v (%v now)\n", c.RemoteAddr(), player.Data.LogNameEnter(), l.counter.Count())
}

func (l LogCountResponder[P]) PlayerLeft(c *websocket.Conn, player *BinaryPlayer[P]) {
	l.LogResponder.PlayerLeft(c, player)
	log.Printf("-[%v] %v (%v now)\n", c.RemoteAddr(), player.Data.LogNameLeave(), l.counter.Count())
}
