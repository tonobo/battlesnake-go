package main // import "github.com/tonobo/battlesnake-go"

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/joonazan/vec2"
)

type Board struct {
	Y      int      `json:"height"`
	X      int      `json:"width"`
	Snakes []*Snake `json:"snakes"`
	Food   []*Food  `json:"food"`

	Me        *Snake   `json:"-"`
	Request   *Request `json:"-"`
	StepLimit int
	debug     bool

	vmap               Map      `json:"-"`
	FakeTargets        []Target `json:"-"`
	boardScoreMap      map[string]int
	boardSquareMap     map[string]int
	disableFakeTargets bool
	disableFoodRoutes  bool
}

func (b *Board) LogFile() io.Writer {
	if b.debug {
		return os.Stdout
	}
	f, err := os.OpenFile(fmt.Sprintf("/var/log/snake-%s-%s.log",
		b.Me.Name,
		b.Request.Game.ID),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	return f
}

func (b *Board) RequestLogFile() io.Writer {
	if b.debug {
		return ioutil.Discard
	}
	f, err := os.OpenFile(fmt.Sprintf("/var/log/access-snake-%s-%s.log",
		b.Me.Name,
		b.Request.Game.ID),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	return f
}

func (b *Board) Target(vec vec2.Vector) Target {
	return b.vmap[int(vec.X)][int(vec.Y)]
}

func (b *Board) BumpScoreAround(vec vec2.Vector, score int) {
	for _, direction := range Direction2Vector {
		next := vec.Minus(direction)
		if b.Outside(next) {
			continue
		}
		b.Target(vec).BumpScore(score)
	}
}

func (b *Board) EmptyConnectedSquaresAround(vec vec2.Vector) int {
	if squares, ok := b.boardSquareMap[fmt.Sprintf("%s", vec)]; ok {
		return squares
	}
	squareMap := make(map[vec2.Vector]struct{}, 0)
	next := vec
	for i := 0; i < b.X*b.Y; i++ {
		for _, direction := range Direction2Vector {
			tmp := next.Minus(direction)
			if _, ok := squareMap[tmp]; ok || b.Blocked(tmp, 0) {
				continue
			}
			squareMap[tmp] = struct{}{}
			next = tmp
		}
	}
	b.boardSquareMap[fmt.Sprintf("%s", vec)] = len(squareMap)
	return len(squareMap)
}

func (b *Board) SnakesAround(vec vec2.Vector, steps int) []*Snake {
	snakes := []*Snake{}
	for _, direction := range Direction2Vector {
		next := vec.Minus(direction)
		if b.Outside(next) {
			continue
		}
		// Check only for snake heads
		if sp := b.SnakeOn(next, steps); sp != nil && (sp.IsHead || sp.IsTail) && sp.Snake.Me != nil {
			snakes = append(snakes, sp.Snake)
		}
	}
	return snakes
}

func (b *Board) FoodAround(vec vec2.Vector) bool {
	for _, direction := range Direction2Vector {
		next := vec.Minus(direction)
		if b.Outside(next) {
			continue
		}
		// Check only for snake heads
		if sp := b.FoodOn(next); sp != nil {
			return true
		}
	}
	return false
}

func shuffle(slice []*Empty) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for len(slice) > 0 {
		n := len(slice)
		randIndex := r.Intn(n)
		slice[n-1], slice[randIndex] = slice[randIndex], slice[n-1]
		slice = slice[:n-1]
	}
}

func (b *Board) VMap() Map {
	if b.vmap != nil {
		return b.vmap
	}
	fakeTargets := make([]*Empty, 0)
	minPercent := (100.0 - float64(FakeTargetsScope)) / 2.0
	maxPercent := 100.0 - float64(minPercent)
	b.vmap = make([][]Target, b.X)
	for x := 0; x < b.X; x++ {
		b.vmap[x] = make([]Target, b.Y)
		for y := 0; y < b.Y; y++ {
			ep := &Empty{Point{X: x, Y: y, Board: b}, false}
			percentY := float64(y) / float64(b.Y) * 100.0
			percentX := float64(x) / float64(b.X) * 100.0
			if percentX >= minPercent && percentX <= maxPercent &&
				percentY >= minPercent && percentY <= maxPercent {
				fakeTargets = append(fakeTargets, ep)
			}
			b.vmap[x][y] = ep
		}
	}
	for _, food := range b.Food {
		food.Board = b
		b.vmap[food.X][food.Y] = food
	}
	for _, snake := range b.Snakes {
		snakeLen := len(snake.Body)
		for i, point := range snake.Body {
			//for i, point := range snake.Body {
			point.Board = b
			point.EvictOnStep = snakeLen - i
			point.Snake = snake
			if i == 0 {
				point.IsHead = true
			}
			b.vmap[point.X][point.Y] = point
		}
		snake.Body[len(snake.Body)-1].IsTail = true
	}
	shuffle(fakeTargets)
	for i := 0; i < MaxFakeTargets; i++ {
		t := fakeTargets[i]
		t.FakeTarget = true
		b.FakeTargets = append(b.FakeTargets, t)
	}
	return b.vmap
}

func (b *Board) Outside(vec vec2.Vector) bool {
	if vec.X > float64(b.X-1) || vec.X < 0.0 {
		return true
	}
	if vec.Y > float64(b.Y-1) || vec.Y < 0.0 {
		return true
	}
	return false
}

func (b *Board) SnakeOn(vec vec2.Vector, steps int) *SnakePoint {
	if target := b.Target(vec); target != nil {
		if sp, ok := target.(*SnakePoint); ok && sp.EvictOnStep > steps {
			return sp
		}
	}
	return nil
}

func (b *Board) FoodOn(vec vec2.Vector) *Food {
	if target := b.Target(vec); target != nil {
		if f, ok := target.(*Food); ok {
			return f
		}
	}
	return nil
}

func (b *Board) Blocked(vec vec2.Vector, steps int) bool {
	if b.Outside(vec) {
		return true
	}
	if b.SnakeOn(vec, steps) != nil {
		return true
	}
	sp, ok := b.Target(vec).(*SnakePoint)
	if ok && sp.IsTail && sp.Snake.Enemy() && b.FoodAround(sp.Snake.Body[0].Vec()) {
		return true
	}
	if ok && sp.IsTail && !sp.Snake.Enemy() && len(sp.Snake.Body) < SnakeMinLenth {
		return true
	}
	return false
}

func (b *Board) OtherSnakesRoutes() Routes {
	r := Routes{}
	for _, snake := range b.Snakes {
		if snake.Me == nil {
			continue
		}
		for i, food := range b.Food {
			route := &Route{
				ID:           i,
				Enemy:        true,
				From:         snake.Body[0],
				To:           food,
				Board:        b,
				stepRegister: make(map[vec2.Vector]struct{}),
				Steps:        Movements{},
			}
			route.Resolve()
			r = append(r, route)
		}
		b.BumpScoreAround(snake.Body[0].Vec(), EnemyHeadBumpScore)
	}
	return r
}

func (b *Board) FoodRoutes() Routes {
	r := make(Routes, len(b.Food))
	for i, food := range b.Food {
		route := &Route{
			ID:           i,
			From:         b.Me.Body[0],
			To:           food,
			Board:        b,
			stepRegister: make(map[vec2.Vector]struct{}),
			Steps:        Movements{},
		}
		route.Resolve()
		route.Print()
		r[i] = route
	}
	return r
}

func (b *Board) SnakeHeadRoutes() Routes {
	r := make(Routes, 0)
	for i, snake := range b.Snakes {
		route := &Route{
			ID:           i + 4000,
			From:         b.Me.Body[0],
			To:           snake.Body[0],
			Board:        b,
			stepRegister: make(map[vec2.Vector]struct{}),
			Steps:        Movements{},
		}
		route.Resolve()
		route.Print()
		r = append(r, route)
	}
	return r
}

func (b *Board) CornerRoutes() Routes {
	r := make(Routes, 0)
	for i, target := range []Target{b.vmap[1][1], b.vmap[1][b.Y-2],
		b.vmap[b.X-2][1], b.vmap[b.X-2][b.Y-2]} {
		route := &Route{
			ID:           i + 6000,
			From:         b.Me.Body[0],
			To:           target,
			Board:        b,
			stepRegister: make(map[vec2.Vector]struct{}),
			Steps:        Movements{},
		}
		route.Resolve()
		route.Print()
		r = append(r, route)
	}
	return r
}

func (b *Board) FakeRoutes() Routes {
	r := make(Routes, len(b.FakeTargets))
	for i, target := range b.FakeTargets {
		route := &Route{
			ID:           i + 2000,
			From:         b.Me.Body[0],
			To:           target,
			Board:        b,
			stepRegister: make(map[vec2.Vector]struct{}),
			Steps:        Movements{},
		}
		route.Resolve()
		route.Print()
		r[i] = route
	}
	return r
}

func (b *Board) Routes() Routes {
	b.OtherSnakesRoutes()
	r := Routes{}
	r = append(r, b.FoodRoutes()...)
	//if len(b.Me.Body) > 3 {
	//	r = append(r, b.TailRoute())
	//}
	if !b.disableFakeTargets {
		if b.Me.Health > SnakeHealthCritical {
			r = append(r, b.FakeRoutes()...)
			r = append(r, b.CornerRoutes()...)
			//	r = append(r, b.SnakeHeadRoutes()...)
		}
	}
	for _, route := range r {
		if !route.Unresolved {
			break
		}
	}
	sort.Sort(r)
	r[0].Print()
	return r
}

func (b *Board) TailRoute() *Route {
	route := &Route{
		ID:           100,
		From:         b.Me.Body[0],
		To:           b.Me.Body[len(b.Me.Body)-1],
		Board:        b,
		stepRegister: make(map[vec2.Vector]struct{}),
		Steps:        Movements{},
		TailRoute:    true,
	}
	route.Resolve()
	route.Print()
	return route
}

func (b *Board) Move() string {
	if b.Me == nil {
		return "up"
	}

	route := b.Routes()[0]
	PrintGrid(b.LogFile(), b.VMap())

	return route.Steps[0].Direction
}
