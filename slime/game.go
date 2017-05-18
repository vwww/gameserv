package slime

import (
	"math/rand"
	"time"
)

const PHYS_FPS = 50
const NETW_FPS = 25
const PHYS_PER_STATE = PHYS_FPS / NETW_FPS

const RAD_PL = 0.1
const RAD_BALL = 0.03

const NET_W = 0.02
const NET_H = 0.175

const BALL_POST_COLLISION_VEL_X_MAX = 0.9375
const BALL_POST_COLLISION_VEL_Y_MAX = 1.375
const BALL_GRAV_ACCEL = 3.125

const PL_SPEED_X = 0.5
const PL_VEL_JUMP = 1.9375
const PL_GRAV_ACCEL = 6.25

type Ball struct {
	MoveState
}

type Game struct {
	P1, P2 *Player
	B      Ball
}

func NewGame(p1, p2 *Player) Game {
	return Game{
		P1: p1,
		P2: p2,
	}
}

func (g *Game) StartRound(p1First bool) {
	g.P1.O.X = 0.45
	g.P1.O.Y = 0
	g.P1.V.X = 0
	g.P1.V.Y = 0
	g.P2.O.X = 1.55
	g.P2.O.Y = 0
	g.P2.V.X = 0
	g.P2.V.Y = 0

	if p1First {
		g.B.O.X = g.P1.O.X
	} else {
		g.B.O.X = g.P2.O.X
	}
	g.B.O.Y = 0.4
	g.B.V.X = 0
	g.B.V.Y = 0
}

func clamp(f *float64, min, max float64) bool {
	if *f < min {
		*f = min
		return true
	} else if *f > max {
		*f = max
		return true
	}
	return false
}

func clampAbs(f *float64, magnitude float64) bool {
	return clamp(f, -magnitude, +magnitude)
}

func moveBallCollide(b *Ball, p *Player) {
	const COLLISION_DIST = RAD_PL + RAD_BALL
	// COLLISION_FACTOR = 2 / (mB/mP + 1)
	//  player mass >> ball mass
	//   -> mB/mP = 0
	const COLLISION_FACTOR = 2

	// difference in position
	dx := b.O.Sub(p.O)
	l := dx.LengthSquared()
	if l > COLLISION_DIST*COLLISION_DIST {
		return
	}

	// move out of the bounding box
	l = dx.Length() // supposedly more accurate than math.Sqrt(l)
	b.O = p.O.Add(dx.Mul(COLLISION_DIST / l * 1.01))

	// elastic collision
	dv := b.V.Sub(p.V)
	b.V = b.V.Sub(dx.Mul(COLLISION_FACTOR * (dx.Dot(dv) / l) / l))

	// limit velocity components
	clampAbs(&b.V.X, BALL_POST_COLLISION_VEL_X_MAX)
	clampAbs(&b.V.Y, BALL_POST_COLLISION_VEL_Y_MAX)
}

func moveBallCollideNet(b *Ball) {
	const NET_HALFW = NET_W / 2 // half-width
	const L = 1 - NET_HALFW
	const R = 1 + NET_HALFW

	// fast bounding box check
	if b.O.Y-RAD_BALL >= NET_H ||
		b.O.X+RAD_BALL <= L ||
		b.O.X-RAD_BALL >= R {
		return
	}

	closest := b.O
	clamp(&closest.X, L, R)
	clamp(&closest.Y, 0, NET_H)

	normal := b.O.Sub(closest)

	if closest.X == b.O.X && closest.Y == b.O.Y {
		normal.X = -normal.X
		normal.Y = -normal.Y
		goto INSIDE
	}

	if normal.LengthSquared() > RAD_BALL*RAD_BALL {
		return
	}
INSIDE:

	l := normal.Length()

	// move out of the bounding box
	b.O = b.O.Add(normal.Mul(RAD_BALL / l * 1.01))

	// elastic collision (net has infinite mass)
	b.V = b.V.Sub(normal.Mul(2 * (normal.Dot(b.V) / l) / l))
}

func moveBall(g *Game) bool {
	hitGround := false

	b := &g.B
	// update positions
	b.V.Y -= BALL_GRAV_ACCEL / PHYS_FPS
	b.O.X += b.V.X / PHYS_FPS
	b.O.Y += b.V.Y / PHYS_FPS

	// collide with players
	moveBallCollide(b, g.P1)
	moveBallCollide(b, g.P2)

	// collide with net
	moveBallCollideNet(b)

	// constrain xRAD_BALL
	if clamp(&b.O.X, RAD_BALL, 2-RAD_BALL) {
		b.V.X = -b.V.X
	}

	// constrain y (check if floor is hit, ignore upper bound)
	if b.O.Y < RAD_BALL {
		b.O.Y = RAD_BALL
		hitGround = true
	}

	return hitGround
}

func movePlayer(p *Player, left bool) {
	// simple horizontal movements
	if p.L != p.R {
		if p.L == left {
			p.V.X = -PL_SPEED_X
		} else {
			p.V.X = +PL_SPEED_X
		}
	} else {
		p.V.X = 0
	}
	// can jump on floor
	if p.U && p.O.Y == 0 {
		p.V.Y += PL_VEL_JUMP
	}

	L, R := RAD_PL, 1.0-RAD_PL-NET_W/2
	if !left {
		L, R = 1.0+RAD_PL+NET_W/2, 2.0-RAD_PL
	}

	// Move X
	p.O.X += p.V.X / PHYS_FPS
	if clamp(&p.O.X, L, R) {
		p.V.X = 0
	}

	// Move Y
	if p.O.Y != 0 || p.V.Y != 0 {
		p.V.Y -= PL_GRAV_ACCEL / PHYS_FPS
		p.O.Y += p.V.Y / PHYS_FPS
		if p.O.Y <= 0 {
			p.O.Y = 0
			p.V.Y = 0 // stick to ground
		} else if p.O.Y > 1 {
			p.O.Y = 1
		}
	}
}

func (g *Game) PhysicsFrame(winner *int) {
	// Move players first
	movePlayer(g.P1, true)
	movePlayer(g.P2, false)
	// Move ball if necessary
	if *winner == 0 {
		if moveBall(g) {
			// Check winner by position of ball
			if g.B.O.X < 1 {
				*winner = 2
			} else {
				*winner = 1
			}
		}
	}
}

func (g *Game) Run() {
	g.P1.SendEnter(g.P2.Name, g.P2.Color)
	g.P2.SendEnter(g.P1.Name, g.P1.Color)

	t := time.Now()
	winner := 3
	p1First := (rand.Intn(2) == 0)
	intermissionEnd := time.Time{}
	pingDivider := 0

GAME_LOOP:
	for {
		switch {
		case g.P1.Stopped:
			g.P2.SendLeave()
			break GAME_LOOP
		case g.P2.Stopped:
			g.P1.SendLeave()
			break GAME_LOOP
		}

		n := time.Now()

		for n.After(t) {
			oldWinner := winner
			for i := 0; i < PHYS_PER_STATE; i++ {
				g.PhysicsFrame(&winner)
			}

			g.P1.SendState(transformState(g.P1, g.P2, g.B.MoveState, true))
			g.P2.SendState(transformState(g.P1, g.P2, g.B.MoveState, false))
			// Send ping times when we have measurements for both
			p1Ping := g.P1.Ping
			if p1Ping != -1 {
				p2Ping := g.P2.Ping
				if p2Ping != -1 {
					g.P1.SendPingTimes(p1Ping, p2Ping)
					g.P2.SendPingTimes(p2Ping, p1Ping)
				}
			}

			if winner != 0 {
				if oldWinner == 0 {
					g.P1.SendEndRound(winner == 1)
					g.P2.SendEndRound(winner == 2)
					intermissionEnd = n.Add(750 * time.Millisecond)
				} else if n.After(intermissionEnd) {
					winner = 0
					p1First = !p1First
					g.P1.SendNextRound(p1First)
					g.P2.SendNextRound(!p1First)
					g.StartRound(p1First)
				}
			}
			t = t.Add(1000 / NETW_FPS * time.Millisecond)
		}

		if pingDivider == 0 {
			g.P1.SendPing()
			g.P2.SendPing()
			pingDivider = 6
		} else {
			pingDivider--
		}

		time.Sleep(40 * time.Millisecond)
	}
}
