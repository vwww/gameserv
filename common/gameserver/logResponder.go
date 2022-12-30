package gameserver

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type LogResponder[P any] struct{ Responder[P] }

func assertInterface_LogResponder[P any]() { var _ Responder[P] = LogResponder[P]{} }

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

type LogNamerPointer[B any] interface {
	*B
	LogNamer
}

type LogCountResponder[B any, P LogNamerPointer[B]] struct {
	LogResponder[P]
	counter Counter
}

func assertInterface_LogCountResponder[B any, P LogNamerPointer[B]]() {
	var _ Responder[P] = LogCountResponder[B, P]{}
}

func NewLogCountResponder[B any, P LogNamerPointer[B]](r Responder[P], counter Counter) LogCountResponder[B, P] {
	return LogCountResponder[B, P]{
		NewLogResponder(r),
		counter,
	}
}

func (l LogCountResponder[B, P]) PlayerJoined(c *websocket.Conn, player *BinaryPlayer[P]) {
	l.LogResponder.PlayerJoined(c, player)
	log.Printf("+[%v] %v (%v now)\n", c.RemoteAddr(), player.Data.LogNameEnter(), l.counter.Count())
}

func (l LogCountResponder[B, P]) PlayerLeft(c *websocket.Conn, player *BinaryPlayer[P]) {
	l.LogResponder.PlayerLeft(c, player)
	log.Printf("-[%v] %v (%v now)\n", c.RemoteAddr(), player.Data.LogNameLeave(), l.counter.Count())
}
