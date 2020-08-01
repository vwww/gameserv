package duel

import (
	"sync"
	"time"
)

// Timing constants
const (
	// Physics frames per second
	PHYS_FPS = 50
	// Network world states per second
	NETW_FPS = 25
	// Interval of physics frames
	PHYS_TIME = time.Second / PHYS_FPS
	// Interval of world state updates
	NETW_TIME = time.Second / NETW_FPS
	// Interval of pings
	PING_TIME = 250 * time.Millisecond
)

// Limits
const (
	// Maximum number of players
	MAX_PL = 256
	// Target number of bots
	BOT_BALANCE = 16
)

// A server should have one Game instance,
// and execute Game.Run() in a new goroutine.
type Game struct {
	players [MAX_PL]Player
	pLock   sync.Mutex

	pCount     int // current number of players
	pCountLock sync.Mutex

	gameStart      time.Time
	lastPhysics    time.Time
	lastWorldState time.Time
	nextPing       time.Time
}

func NewGame() *Game {
	var g Game
	for i := 0; i < BOT_BALANCE; i++ {
		g.players[i].InitBot()
	}
	return &g
}

// AddPlayer adds a remotely-controlled player to the game and returns a Client,
// or nil on failure.
func (g *Game) AddPlayer(name []byte, col uint8) *Client {
	g.pLock.Lock()
	defer g.pLock.Unlock()

	for i := range g.players {
		p := &g.players[i]
		if !p.IsValid || p.Client == nil {
			p.InitPlayer(name, col)

			p.Client = newClient(g, i)

			p.Client.SendB(MsgWelcome(i))
			msg := PrepareMessage(MsgEnter(i, p.Color, 0, 0, 0, 0, p.Name))
			for j := range g.players {
				pp := &g.players[j]
				if i == j || !pp.IsValid {
					continue
				} else if pp.Client != nil {
					p.Client.SendB(MsgEnter(
						j, pp.Color,
						pp.Kills, pp.Deaths, pp.Combo, pp.Score,
						pp.Name,
					))
					pp.Client.Send(msg)
				} else {
					p.Client.SendB(MsgEnterBot(
						j, pp.Color,
						pp.Kills, pp.Deaths, pp.Combo, pp.Score,
						pp.Name,
					))
				}
			}

			g.pCountLock.Lock()
			defer g.pCountLock.Unlock()
			g.pCount++

			return p.Client
		}
	}

	return nil
}

// DelPlayer removes a player from the game.
func (g *Game) DelPlayer(cn int) {
	g.pLock.Lock()
	defer g.pLock.Unlock()

	p := &g.players[cn]
	if p.IsValid {
		p.Reset()
	}

	var msg []byte
	g.pCountLock.Lock()
	defer g.pCountLock.Unlock()
	g.pCount--

	if g.pCount < BOT_BALANCE {
		// Replace with bot
		p.InitBot()
		msg = MsgEnterBot(cn, p.Color, 0, 0, 0, 0, p.Name)
	} else {
		// Remove player
		msg = MsgLeave(cn)
	}

	g.Broadcast(msg)
}

// Broadcast sends a message to all players
func (g *Game) Broadcast(msg []byte) {
	pm := PrepareMessage(msg)
	for i := range g.players {
		p := &g.players[i]
		if p.IsValid && p.Client != nil {
			p.Client.Send(pm)
		}
	}
}

// serverslice periodically runs, and runs needed processing for the game.
func (g *Game) serverslice() {
	g.pLock.Lock()
	defer g.pLock.Unlock()

	now := time.Now()

	// Apply physics
	if now.After(g.lastPhysics) {
		g.PhysicsFrame()
		g.lastPhysics = g.lastPhysics.Add(PHYS_TIME)
	}

	// Send world state
	if now.After(g.lastWorldState) {
		g.Broadcast(buildWorldState(g))
		g.lastWorldState = g.lastWorldState.Add(NETW_TIME)
	}

	// Send pings and ping results
	if now.After(g.nextPing) {
		g.Broadcast(MsgPing())
		g.nextPing = now.Add(PING_TIME)
	}
}

// Run is a loop that runs the game forever.
func (g *Game) Run() {
	// timers
	now := time.Now()
	g.gameStart = now
	g.lastPhysics = now
	g.lastWorldState = now
	g.nextPing = now
	for {
		g.serverslice()
		time.Sleep(10 * time.Millisecond)
	}
}
