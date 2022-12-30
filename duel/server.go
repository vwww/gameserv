// Package duel implements the logic for the Duel server.
package duel

import (
	"victorz.ca/gameserv/common/gameserver"

	"github.com/gorilla/websocket"
)

// Server is a Duel game server.
type Server struct {
	gameserver.Responder[*Client]
	*gameserver.GameServerCount[Client]
	*Game
}

// NewServer makes a new game server.
func NewServer() Server {
	const sendBufSize = 300 // enough for at least 2 seconds

	var s Server
	s.Game = NewGame()

	r := gameserver.DefaultResponder[Client]()
	r = gameserver.NewLogCountResponder(r, &s)
	s.Responder = r
	s.GameServerCount = gameserver.NewGameServerCount[Client](&s, sendBufSize)
	return s
}

// Run runs the game server. It should normally be called in its
// own goroutine.
func (s *Server) Run() {
	// s.Game.Run()
}

func (s *Server) PlayerInit(c *websocket.Conn) *Client {
	return s.AddPlayer(processHello(c))
}

func (s *Server) PlayerJoined(c *websocket.Conn, player *gameserver.BinaryPlayer[*Client]) {
	s.Responder.PlayerJoined(c, player)

	player.Data.Conn = c
}

func (s *Server) PlayerLeft(c *websocket.Conn, player *gameserver.BinaryPlayer[*Client]) {
	player.Data.Close()

	s.Responder.PlayerLeft(c, player)
}

func (s *Server) MessageReceived(player *gameserver.BinaryPlayer[*Client], msg []byte) {
	s.Responder.MessageReceived(player, msg)

	Recv(player.Data, msg)
}
