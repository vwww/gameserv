package gameserver

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Msg is a websocket message.
type Msg struct {
	MsgType int
	Payload []byte
}

// Player represents a connected client.
type Player struct {
	Data interface{}
	Stop chan struct{}

	recv     func(Msg)
	sendBuf  chan Msg
	stopOnce sync.Once
}

// NewPlayer makes a Player with the embedded data, receive callback, and send buffer size.
func NewPlayer(data interface{}, recv func(Msg), sendBufSize uint) Player {
	return Player{
		data,
		make(chan struct{}),

		recv,
		make(chan Msg, sendBufSize),
		sync.Once{},
	}
}

// Send enqueues an outgoing message, or
// on failure, closes the Player.
func (p *Player) Send(msg Msg) {
	select {
	case p.sendBuf <- msg:
	default:
		// queue overflow
		p.Close()
	}
}

// Close marks the player as "stopped" by closing the send and stop channels.
// Close must not be called multiple times.
func (p *Player) Close() {
	close(p.Stop)
	close(p.sendBuf)
}

// BinaryPlayer is an adapter for Player, which sends binary messages
// and ignores incoming message types.
type BinaryPlayer struct {
	Player

	Recv func([]byte)
}

// NewBinaryPlayer makes a BinaryPlayer.
func NewBinaryPlayer(data interface{}, recv func([]byte), sendBufSize uint) *BinaryPlayer {
	var p BinaryPlayer
	p.Recv = recv
	p.Player = NewPlayer(data, func(m Msg) { p.Recv(m.Payload) }, sendBufSize)
	return &p
}

// Send sends the byte slice as a binary message over the websocket.
func (p *BinaryPlayer) Send(b []byte) {
	p.Player.Send(Msg{websocket.BinaryMessage, b})
}
