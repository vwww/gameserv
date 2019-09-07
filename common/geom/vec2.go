package geom

import (
	"math"
)

// Vec2 represents a 2-dimensional vector.
type Vec2 struct {
	X, Y float64
}

// Add returns the result of adding another Vec2.
func (v Vec2) Add(w Vec2) Vec2 { return Vec2{v.X + w.X, v.Y + w.Y} }

// Sub returns the result of subtracting another Vec2.
func (v Vec2) Sub(w Vec2) Vec2 { return Vec2{v.X - w.X, v.Y - w.Y} }

// Mul multiplies the vector by a scalar.
func (v Vec2) Mul(f float64) Vec2 { return Vec2{v.X * f, v.Y * f} }

// Div divides the vector by a scalar.
func (v Vec2) Div(f float64) Vec2 { return Vec2{v.X / f, v.Y / f} }

// Dot returns the dot product with another Vec2.
func (v Vec2) Dot(w Vec2) float64 { return v.X*w.X + v.Y*w.Y }

// Length returns the magnitude (Euclidean length).
func (v Vec2) Length() float64 { return math.Hypot(v.X, v.Y) }

// LengthSquared returns the magnitude multiplied by itself.
func (v Vec2) LengthSquared() float64 { return v.Dot(v) }

// Normalize returns a unit vector in the same direction.
// The result is undefined if the length is zero.
func (v Vec2) Normalize() Vec2 { return v.Div(v.Length()) }
