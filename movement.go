package main // import "github.com/tonobo/battlesnake-go"

import "github.com/joonazan/vec2"

type Movement struct {
	Direction string
	Magnitude float64
	Target    vec2.Vector
}

func (m *Movement) X() int {
	return int(m.Target.X)
}

func (m *Movement) Y() int {
	return int(m.Target.Y)
}

type Movements []*Movement

func (p Movements) Len() int           { return len(p) }
func (p Movements) Less(i, j int) bool { return p[i].Magnitude < p[j].Magnitude }
func (p Movements) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
