package gameserver

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// GameServerCount extends BaseGameServer by counting the number of players.
type GameServerCount struct {
	BaseGameServer
	Responder

	count     uint // current number of players
	countLock sync.RWMutex
}

// NewGameServerCount makes a new GameServerCount for the specified responder and send buffer size.
func NewGameServerCount(r Responder, sendBufSize uint) *GameServerCount {
	g := GameServerCount{
		BaseGameServer: BaseGameServer{
			nil,
			sendBufSize,
		},
		Responder: r,
	}
	g.BaseGameServer.Responder = gameServerCountImpl{&g}
	return &g
}

// Count returns the current number of players.
func (g *GameServerCount) Count() uint {
	g.countLock.RLock()
	defer g.countLock.RUnlock()
	return g.count
}

// HandleNum responds to the HTTP request by writing the current number of players.
func (g *GameServerCount) HandleNum(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v", g.Count())
}

type gameServerCountImpl struct{ *GameServerCount }

var _ Responder = gameServerCountImpl{}

func (g gameServerCountImpl) PlayerJoined(c *websocket.Conn, player *BinaryPlayer) {
	g.countLock.Lock()
	g.count++
	g.countLock.Unlock()

	g.Responder.PlayerJoined(c, player)
}

func (g gameServerCountImpl) PlayerLeft(c *websocket.Conn, player *BinaryPlayer) {
	g.countLock.Lock()
	g.count--
	g.countLock.Unlock()

	g.Responder.PlayerLeft(c, player)
}
