package main // import "github.com/tonobo/battlesnake-go"

import (
	"fmt"

	"github.com/joonazan/vec2"
)

type Map [][]Target
type Target interface {
	Vec() vec2.Vector
	Type() string
	BumpScore(...int)
	Score() int
}

type Point struct {
	Y     int `json:"y"`
	X     int `json:"x"`
	score int
	Index int
	Board *Board `json:"-"`
}

func (p *Point) Vec() vec2.Vector {
	return vec2.Vector{float64(p.X), float64(p.Y)}
}

func (p *Point) BumpScore(s ...int) {
	if len(s) > 0 {
		p.Board.boardScoreMap[fmt.Sprintf("%s", p.Vec())] += s[0]
	} else {
		p.Board.boardScoreMap[fmt.Sprintf("%s", p.Vec())] += 1
	}
}

func (p *Point) Score() int {
	i, _ := p.Board.boardScoreMap[fmt.Sprintf("%s", p.Vec())]
	return i
}

type Empty struct {
	Point
	FakeTarget bool
}

func (f *Empty) Type() string {
	return "empty"
}

type Food struct {
	Point
}

func (f *Food) Type() string {
	return "food"
}
