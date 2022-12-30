package gameserver

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// GameServerCount extends BaseGameServer by counting the number of players.
type GameServerCount[P any] struct {
	BaseGameServer[P]

	count     uint // current number of players
	countLock sync.RWMutex
}

// NewGameServerCount makes a new GameServerCount for the specified responder and send buffer size.
func NewGameServerCount[P any](r Responder[*P], sendBufSize uint) *GameServerCount[P] {
	g := GameServerCount[P]{
		BaseGameServer: BaseGameServer[P]{
			nil,
			sendBufSize,
		},
	}
	g.BaseGameServer.Responder = gameServerCountImpl[P]{r, &g}
	return &g
}

// Count returns the current number of players.
func (g *GameServerCount[P]) Count() uint {
	g.countLock.RLock()
	defer g.countLock.RUnlock()
	return g.count
}

// HandleNum responds to the HTTP request by writing the current number of players.
func (g *GameServerCount[P]) HandleNum(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v", g.Count())
}

type gameServerCountImpl[P any] struct {
	Responder[*P]
	*GameServerCount[P]
}

func assertInterface_gameServerCountImpl[P any]() { var _ Responder[*P] = gameServerCountImpl[P]{} }

func (g gameServerCountImpl[P]) PlayerJoined(c *websocket.Conn, player *BinaryPlayer[*P]) {
	g.countLock.Lock()
	g.count++
	g.countLock.Unlock()

	g.Responder.PlayerJoined(c, player)
}

func (g gameServerCountImpl[P]) PlayerLeft(c *websocket.Conn, player *BinaryPlayer[*P]) {
	g.countLock.Lock()
	g.count--
	g.countLock.Unlock()

	g.Responder.PlayerLeft(c, player)
}
