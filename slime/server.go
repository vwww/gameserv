// Package slime implements the logic for the Slime Volleyball Multiplayer server.
package slime

import (
	"victorz.ca/gameserv/common/gameserver"

	"github.com/gorilla/websocket"
)

type matchReq struct {
	p      *Player
	result chan struct{}
}

// Server is a Slime Volleyball Multiplayer game server.
type Server struct {
	gameserver.Responder[*Player]
	*gameserver.GameServerCount[Player]
	matcher chan matchReq
}

// NewServer makes a new game server.
func NewServer() Server {
	const sendBufSize = 70 // enough for at least 2 seconds

	var s Server
	s.matcher = make(chan matchReq)
	r := gameserver.DefaultResponder[Player]()
	r = gameserver.NewLogCountResponder(r, &s)
	s.Responder = r
	s.GameServerCount = gameserver.NewGameServerCount[Player](&s, sendBufSize)
	return s
}

// Run does nothing and immediately returns, as slime volleyball multiplayer
// has no tasks to run in the background.
func (s *Server) Run() {
	// does nothing
}

func (s *Server) PlayerInit(c *websocket.Conn) *Player {
	return processHello(c)
}

func (s *Server) PlayerJoined(c *websocket.Conn, player *gameserver.BinaryPlayer[*Player]) {
	s.Responder.PlayerJoined(c, player)

	player.Data.Send = player.Send
	go playMatches(player.Data, s.matcher)
}

func (s *Server) PlayerLeft(c *websocket.Conn, player *gameserver.BinaryPlayer[*Player]) {
	player.Data.Close()

	s.Responder.PlayerLeft(c, player)
}

func (s *Server) MessageReceived(player *gameserver.BinaryPlayer[*Player], msg []byte) {
	s.Responder.MessageReceived(player, msg)

	player.Data.Recv(msg)
}

func processHello(c *websocket.Conn) *Player {
	mt, h, err := c.ReadMessage()

	if mt != websocket.BinaryMessage || err != nil || len(h) < 3 {
		return nil
	}

	name := h[3:]
	col := int(h[0])<<16 | int(h[1])<<8 | int(h[2])

	return NewPlayer(name, col)
}

func playMatches(p *Player, matcher chan matchReq) {
	p.SendWelcome()
	for {
		m := matchReq{p, make(chan struct{})}
		select {
		case <-p.Stop:
			return
		case matcher <- m:
			// wait for game to end
			<-m.result
		case other := <-matcher:
			m.result = nil // free unused chan
			g := NewGame(p, other.p)
			g.Run()
			other.result <- struct{}{}
		}
	}
}
