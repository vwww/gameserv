package gameserver

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// Responder is an interface that handles GameServer events.
type Responder[P any] interface {
	PlayerConnected(r *http.Request)
	PlayerUpgradeFail(r *http.Request, err error)
	PlayerUpgradeSuccess(r *http.Request, c *websocket.Conn)
	PlayerInit(c *websocket.Conn) P
	PlayerJoined(c *websocket.Conn, player *BinaryPlayer[P])
	PlayerLeft(c *websocket.Conn, player *BinaryPlayer[P])
	MessageReceived(player *BinaryPlayer[P], msg []byte)
}

type Tester Responder[func(a, b int) string]

type defaultResponder[P any] struct{}

func assertInterface_defaultResponder[P any]() { var _ Responder[*P] = defaultResponder[P]{} }

func (d defaultResponder[P]) PlayerConnected(r *http.Request)                          {}
func (d defaultResponder[P]) PlayerUpgradeFail(r *http.Request, err error)             {}
func (d defaultResponder[P]) PlayerUpgradeSuccess(r *http.Request, c *websocket.Conn)  {}
func (d defaultResponder[P]) PlayerInit(c *websocket.Conn) *P                          { return nil }
func (d defaultResponder[P]) PlayerJoined(c *websocket.Conn, player *BinaryPlayer[*P]) {}
func (d defaultResponder[P]) PlayerLeft(c *websocket.Conn, player *BinaryPlayer[*P])   {}
func (d defaultResponder[P]) MessageReceived(player *BinaryPlayer[*P], msg []byte)     {}

// DefaultResponder creates a Responder whose empty receivers do nothing.
func DefaultResponder[P any]() Responder[*P] {
	return defaultResponder[P]{}
}

// BaseGameServer is a game server that runs on WebSockets.
type BaseGameServer[P any] struct {
	Responder[*P]

	SendBufSize uint
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func reader(c *websocket.Conn, onMsg func(Msg), onError func(error)) {
	for {
		msgType, msg, err := c.ReadMessage()
		if err != nil {
			onError(nil)
			break
		}
		onMsg(Msg{msgType, msg})
	}
}

func writer(c *websocket.Conn, msgChan <-chan Msg) {
	for msg := range msgChan {
		if err := c.WriteMessage(msg.MsgType, msg.Payload); err != nil {
			break
		}
	}
}

// HandlePlayer serves a game client.
func (g *BaseGameServer[P]) HandlePlayer(w http.ResponseWriter, r *http.Request) {
	g.Responder.PlayerConnected(r)
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		g.Responder.PlayerUpgradeFail(r, err)
		return
	}

	defer c.Close()

	g.Responder.PlayerUpgradeSuccess(r, c)

	data := g.Responder.PlayerInit(c)
	if data == nil {
		return
	}

	p := NewBinaryPlayer(
		data,
		nil,
		g.SendBufSize,
	)
	p.Recv = func(msg []byte) { g.Responder.MessageReceived(p, msg) }

	defer g.Responder.PlayerLeft(c, p)
	g.Responder.PlayerJoined(c, p)

	go reader(c, p.Player.recv, func(error) { p.Close() })
	writer(c, p.Player.sendBuf)
}
