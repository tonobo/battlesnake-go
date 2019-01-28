package main // import "github.com/tonobo/battlesnake-go"

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joonazan/vec2"
)

var (
	Direction2Vector = map[string]vec2.Vector{
		"left":  vec2.Vector{1, 0},
		"up":    vec2.Vector{0, 1},
		"right": vec2.Vector{-1, 0},
		"down":  vec2.Vector{0, -1},
	}

	StepLimit                = 50
	UnresolvedScoreBump      = 300
	DirectNeighborScoreBump  = 80
	EnemyFoodBumpScoreOnTail = 100
	EnemyMaxSteps            = 3
	BestMoveSelection        = 1 // how many steps
	UnavailableTailScoreBump = 50
	FoodHealthLimit          = 50
	SnakeHealthCritical      = 20
	BorderBumpScore          = 10
	MaxFakeTargets           = 6
	MinBackrouteSteps        = 15
	FakeTargetsScope         = 70 // how many percent from board
	EnemyHeadBumpScore       = 5
	FoodBumpScore            = 0
	SnakeMinLenth            = 2
	EnemyBumpScore           = 5
	SnakeAroundBump          = 5
	SmallSnakeAroundBump     = 0
	SnakeIDList              = []string{"a", "b", "c", "d", "e", "g", "h", "j", "k"}
)

type Request struct {
	Game  *Game  `json:"game"`
	Turn  int    `json:"turn"`
	Board *Board `json:"board"`
	Self  *Snake `json:"you"`
}

func boolPtr(b bool) *bool {
	return &b
}

func (r *Request) Init() {
	r.Board.StepLimit = StepLimit
	r.Board.Me = r.Self
	r.Board.Request = r
	for x, rows := range r.Board.VMap() {
		for y, target := range rows {
			if y == 0 || x == r.Board.Y || x == 0 || x == r.Board.X {
				target.BumpScore(BorderBumpScore)
			}
		}
	}
	for i, snake := range r.Board.Snakes {
		snake.Board = r.Board
		if r.Self.Body[0].X == snake.Body[0].X && r.Self.Body[0].Y == snake.Body[0].Y {
			continue
		}
		snake.InternalID = SnakeIDList[i]
		snake.Me = r.Self
	}
	r.Self.Board = r.Board
}

type Game struct {
	ID string `json:"id"`
}

func PrintGrid(file io.Writer, grid Map) {
	for i := 0; i < len(grid); i++ {
		for j := 0; j < len(grid[0]); j++ {
			switch s := grid[j][i].(type) {
			case *Food:
				fmt.Fprintf(file, "F")
			case *Empty:
				if s.FakeTarget {
					fmt.Fprintf(file, "X")
				} else {
					fmt.Fprintf(file, "-")
				}
			case *SnakePoint:
				if s.Snake.Me == nil {
					if s.IsHead {
						fmt.Fprintf(file, "M")
					} else {
						fmt.Fprintf(file, "m")
					}
				} else {
					if s.IsHead {
						fmt.Fprintf(file, strings.ToUpper(s.Snake.InternalID))
					} else {
						fmt.Fprintf(file, s.Snake.InternalID)
					}
				}
			default:
				fmt.Fprint(file, "-")
			}
		}
		fmt.Fprint(file, "\n")
	}
	fmt.Fprint(file, "\n")
}

var (
	move = flag.Bool("move", false, "Load move")
)

func main() {
	flag.Parse()
	if *move {
		var j Request
		err := json.NewDecoder(os.Stdin).Decode(&j)
		if err != nil {
			panic(err)
		}
		j.Init()
		j.Board.debug = true
		fmt.Println(j.Board.Move())
		return
	}
	r := gin.Default()

	r.POST("/start", func(c *gin.Context) {
		var json Request
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		json.Init()
		fmt.Printf("Starting game: %s\n", json.Game.ID)
		c.JSON(http.StatusOK, gin.H{"color": "#ff00ff"})
	})

	r.POST("/end", func(c *gin.Context) {
		var j Request
		if err := c.ShouldBindJSON(&j); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		j.Init()
		body, _ := json.Marshal(j)
		fmt.Fprintf(j.Board.RequestLogFile(), "%s\n", body)
		fmt.Printf("End game: %s\n", j.Game.ID)
		c.JSON(http.StatusOK, gin.H{})
	})

	r.POST("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	r.POST("/move", func(c *gin.Context) {
		var j Request
		if err := c.ShouldBindJSON(&j); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		j.Init()
		body, _ := json.Marshal(j)
		fmt.Fprintf(j.Board.RequestLogFile(), "%s\n", body)

		fmt.Printf("Move game: %s\n", j.Game.ID)
		c.JSON(http.StatusOK, gin.H{"move": j.Board.Move()})
	})

	// Listen and serve on 0.0.0.0:8080
	r.Run()
}
