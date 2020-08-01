package gameserver

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// Responder is an interface that handles GameServer events.
type Responder interface {
	PlayerConnected(r *http.Request)
	PlayerUpgradeFail(r *http.Request, err error)
	PlayerUpgradeSuccess(r *http.Request, c *websocket.Conn)
	PlayerInit(c *websocket.Conn) interface{}
	PlayerJoined(c *websocket.Conn, player *BinaryPlayer)
	PlayerLeft(c *websocket.Conn, player *BinaryPlayer)
	MessageReceived(player *BinaryPlayer, msg []byte)
}

type defaultResponder struct{}

var _ Responder = defaultResponder{}

func (d defaultResponder) PlayerConnected(r *http.Request)                         {}
func (d defaultResponder) PlayerUpgradeFail(r *http.Request, err error)            {}
func (d defaultResponder) PlayerUpgradeSuccess(r *http.Request, c *websocket.Conn) {}
func (d defaultResponder) PlayerInit(c *websocket.Conn) interface{}                { return nil }
func (d defaultResponder) PlayerJoined(c *websocket.Conn, player *BinaryPlayer)    {}
func (d defaultResponder) PlayerLeft(c *websocket.Conn, player *BinaryPlayer)      {}
func (d defaultResponder) MessageReceived(player *BinaryPlayer, msg []byte)        {}

// DefaultResponder creates a Responder whose empty receivers do nothing.
func DefaultResponder() Responder {
	return defaultResponder{}
}

// BaseGameServer is a game server that runs on WebSockets.
type BaseGameServer struct {
	Responder Responder

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
func (g *BaseGameServer) HandlePlayer(w http.ResponseWriter, r *http.Request) {
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
