package main // import "github.com/tonobo/battlesnake-go"

import (
	"fmt"
	"sort"

	"github.com/joonazan/vec2"
)

type Route struct {
	From  Target
	To    Target
	Board *Board

	ID            int
	StepCount     int
	Squares       int
	Steps         Movements
	Score         float64
	FieldScore    float64
	Aborted       bool
	Unresolved    bool
	Info          string
	Enemy         bool
	TailReachable *bool
	TailRoute     bool

	stepRegister map[vec2.Vector]struct{}
}

type Routes []*Route

func (p Routes) Len() int           { return len(p) }
func (p Routes) Less(i, j int) bool { return p[i].Score < p[j].Score }
func (p Routes) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type LongestRoutes []*Route

func (p LongestRoutes) Len() int           { return len(p) }
func (p LongestRoutes) Less(i, j int) bool { return p[i].StepCount < p[j].StepCount }
func (p LongestRoutes) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (r *Route) AlreadyUsed(v vec2.Vector) bool {
	_, found := r.stepRegister[v]
	return found
}

func (r *Route) AddStep(m *Movement) {
	if r.Enemy {
		r.Board.VMap()[m.X()][m.Y()].BumpScore()
	}
	r.StepCount++
	r.Steps = append(r.Steps, m)
	r.Score += float64(r.StepCount)
	r.stepRegister[m.Target] = struct{}{}
}

func (r *Route) CheckBackRoute() {
	if !r.Unresolved && !r.Enemy && r.TailReachable == nil {
		backRoute := &Route{
			ID:            r.ID + 1000,
			From:          r.To,
			To:            r.Board.Me.Body[len(r.Board.Me.Body)-1],
			Board:         r.Board,
			stepRegister:  make(map[vec2.Vector]struct{}),
			Steps:         Movements{},
			TailReachable: boolPtr(true),
		}
		backRoute.Resolve()
		if !backRoute.Unresolved || backRoute.StepCount > MinBackrouteSteps {
			r.TailReachable = boolPtr(true)
		} else {
			r.Info += "T"
			r.Score += float64(UnavailableTailScoreBump)
			r.TailReachable = boolPtr(false)
		}
	}
}

func (r *Route) Resolve() {
	defer func() {
		if r.StepCount < r.Board.StepLimit && !r.Unresolved && !r.Enemy {
			r.Board.StepLimit = r.StepCount
		}
		if len(r.Steps) > 0 {
			r.Squares = r.Board.EmptyConnectedSquaresAround(r.Steps[0].Target)
		}
		r.CheckBackRoute()
		r.Score = r.Score / float64(r.StepCount+r.Squares)
	}()
	last := &Movement{Target: r.From.Vec()}
	if r.Enemy && r.Board.FoodAround(last.Target) {
		eb := r.From.(*SnakePoint).Snake.Body
		eb[len(eb)-1].BumpScore(EnemyFoodBumpScoreOnTail)
	}
	for {
		moves := Movements{}
		for direction, vec := range Direction2Vector {
			next := last.Target.Minus(vec)
			movement := &Movement{Direction: direction, Target: next, Magnitude: next.Minus(r.To.Vec()).Length()}
			if r.Board.Blocked(next, r.StepCount) {
				continue
			}
			if r.AlreadyUsed(next) {
				continue
			}
			if !r.Enemy {
				r.Score += float64(r.Board.VMap()[int(next.X)][int(next.Y)].Score())
				r.FieldScore += float64(r.Board.VMap()[int(next.X)][int(next.Y)].Score())
				if r.StepCount < BestMoveSelection {
					for range r.Board.SnakesAround(last.Target, r.StepCount) {
						movement.Magnitude += 1
					}
				}
			}
			moves = append(moves, movement)

		}
		if len(moves) == 0 {
			r.Info += "U"
			r.Score += float64(UnresolvedScoreBump)
			r.Unresolved = true
			return
		}
		sort.Sort(moves)
		last = moves[0]
		r.AddStep(last)
		if !r.Enemy {
			for range r.Board.SnakesAround(last.Target, r.StepCount) {
				r.Info += "S"
				r.Score += float64(SnakeAroundBump)
			}
		}
		if r.StepCount == 1 {
			for _, snake := range r.Board.SnakesAround(last.Target, r.StepCount) {
				// If neighbor snake is larger or equal
				if len(snake.Body) >= len(r.Board.Me.Body) {
					r.Info += "U"
					r.Score += float64(DirectNeighborScoreBump)
					r.Unresolved = true
					return
				}
			}
		}
		if last.Magnitude == 0.0 {
			return
		}
		if r.Enemy {
			for _, step := range r.Steps {
				r.Info += "B"
				r.Board.VMap()[step.X()][step.Y()].BumpScore(EnemyBumpScore)
			}
			if r.StepCount > EnemyMaxSteps {
				return
			}
		}
		if r.StepCount >= StepLimit {
			r.Aborted = true
			return
		}
	}
	return
}

func (r *Route) Print() {
	vec := r.To.Vec()
	tail := "nil"
	if r.TailReachable != nil && *r.TailReachable {
		tail = "true"
	} else if r.TailReachable != nil && !*r.TailReachable {
		tail = "false"
	}
	if r.StepCount > 0 {
		fmt.Fprintf(r.Board.LogFile(),
			"%d. %s: x: %0.f, y: %0.f, distance: %f, step count: %d,"+
				" step: %s%v, score: %0.2f, field_score: %0.2f, aborted: %t"+
				", unresolved: %t, tail_reachable: %s, info: %s, squares: %d\n",
			r.ID, r.To.Type(), vec.X, vec.Y,
			r.From.Vec().Minus(vec).Length(),
			r.StepCount, r.Steps[0].Direction, r.Steps[0].Target, r.Score,
			r.FieldScore, r.Aborted, r.Unresolved, tail, r.Info,
			r.Squares)
		return
	}
	fmt.Fprintf(r.Board.LogFile(),
		"%d. %s: x: %0.f, y: %0.f, distance: %f, step count: %d\n",
		r.ID, r.To.Type(), vec.X, vec.Y,
		r.From.Vec().Minus(vec).Length(),
		r.StepCount)
}
