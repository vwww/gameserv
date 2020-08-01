package slime

import (
	"bytes"
	"fmt"
	"sync"

	"victorz.ca/gameserv/common/geom"
)

// MoveState is the origin (position) and velocity of dynamic entities.
type MoveState struct {
	O, V geom.Vec2
}

// InputState represents client input (which keys on keyboard are pressed).
type InputState struct {
	L, R, U bool
}

// Player represents a connected client.
type Player struct {
	// Inputs
	Name  string
	Color int
	InputState

	// Game State
	MoveState

	Stop     chan struct{}
	stopOnce sync.Once

	Ping int
	RemotePlayer
}

// NewPlayer makes a player with the given name and color.
func NewPlayer(name []byte, col int) *Player {
	p := new(Player)
	p.Name = filterName(name)
	p.Color = filterColor(col)
	p.Ping = -1
	p.Stop = make(chan struct{})
	p.RemotePlayer = newRemotePlayer(p)
	return p
}

// Close marks the player as "stopped" and closes the Stop chan, so
// the Stop chan will no longer block.
func (p *Player) Close() {
	p.stopOnce.Do(func() {
		close(p.Stop)
	})
}

// LogNameEnter returns a name for logging when connecting.
func (p *Player) LogNameEnter() string {
	return fmt.Sprintf("%v #%06x", p.Name, p.Color)
}

// LogNameLeave returns a name for logging when leaving.
func (p *Player) LogNameLeave() string {
	return p.Name
}

// filterName sanitizes a name.
// Invalid characters are removed.
// Names are truncated if they are too long.
// If a name would be blank, a valid name will be returned instead.
func filterName(name []byte) string {
	const MAX_NAME_LEN = 16

	name = bytes.Map(func(r rune) rune {
		switch {
		case (r >= '0' && r <= '9'),
			(r >= 'A' && r <= 'Z'),
			(r >= 'a' && r <= 'z'),
			r == '-',
			r == ' ':
			return r

		default:
			return -1
		}
	}, name)

	name = bytes.TrimSpace(name)
	if len(name) > MAX_NAME_LEN {
		name = name[:MAX_NAME_LEN]
	}

	// Convert to string
	s := "unnamed"
	if len(name) != 0 {
		s = string(name)
	}
	return s
}

// filterColor sanitizes a color value.
func filterColor(c int) int {
	return c & 0xFFFFFF
}
