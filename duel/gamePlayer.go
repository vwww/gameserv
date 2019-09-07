package duel

import (
	"bytes"
	"math"
	"math/rand"
	"sync"

	"victorz.ca/gameserv/common/geom"
)

// Player represents a player, either a remotely-connected client or a local bot.
type Player struct {
	// Inputs
	Name  string
	Color uint8
	D     geom.Vec2 // Destination

	// Game State
	O geom.Vec2 // Origin
	M uint      // Mass
	R float64   // Radius (cuberoot(mass))

	Kills   uint
	Deaths  uint
	Combo   uint
	IsAlive bool

	Score uint

	IsValid bool
	*Client
	BotDivider uint

	sync.Mutex
}

// Reset destroys any resources allocated to this Player object,
// and marks the player as invalid.
func (p *Player) Reset() {
	p.Name = ""
	p.Client = nil
	p.IsValid = false
	p.IsAlive = false
}

func (p *Player) init() {
	p.Kills = 0
	p.Deaths = 0
	p.IsAlive = false
	p.IsValid = true
	p.Score = 0
}

// InitPlayer initializes a remote-controlled Player.
func (p *Player) InitPlayer(name []byte, col uint8) {
	p.init()
	p.Name = filterName(name)
	p.Color = col
}

// InitBot initializes a bot-controlled Player.
func (p *Player) InitBot() {
	p.init()
	p.Name = "" // randomName()
	p.Color = uint8(rand.Intn(0x100))
}

/*
// randomName generates a random player name.
func randomName() string {
	var b [4]byte
	rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
*/

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

func (p *Player) setMass(m uint) {
	if p.M != m {
		p.M = m
		p.R = math.Sqrt(float64(m))
	}
}
