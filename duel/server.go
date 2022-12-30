// Package duel implements the logic for the Duel server.
package duel

import (
	"github.com/gorilla/websocket"
	"victorz.ca/gameserv/common/gameserver"
)

// Server is a Duel game server.
type Server struct {
	*gameserver.GameServerCount
	*Game
}

// NewServer makes a new game server.
func NewServer() Server {
	const sendBufSize = 300 // enough for at least 2 seconds

	var s Server
	s.Game = NewGame()
	s.GameServerCount = gameserver.NewGameServerCount(servImpl{
		gameserver.NewLogCountResponder(gameserver.DefaultResponder(), &s),
		&s,
	}, sendBufSize)
	return s
}

// Run runs the game server. It should normally be called in its
// own goroutine.
func (s *Server) Run() {
	// s.Game.Run()
}

type servImpl struct {
	gameserver.Responder
	server *Server
}

func (s servImpl) PlayerInit(c *websocket.Conn) interface{} {
	return s.server.AddPlayer(processHello(c))
}

func (s servImpl) PlayerJoined(c *websocket.Conn, player *gameserver.BinaryPlayer) {
	s.Responder.PlayerJoined(c, player)

	data := player.Data.(*Player)
	data.Conn = c
}

func (s servImpl) PlayerLeft(c *websocket.Conn, player *gameserver.BinaryPlayer) {
	player.Data.(*Player).Close()

	s.Responder.PlayerLeft(c, player)
}

func (s servImpl) MessageReceived(player *gameserver.BinaryPlayer, msg []byte) {
	s.Responder.MessageReceived(player, msg)

	data := player.Data.(*Player)
	Recv(data.Client, msg)
}
