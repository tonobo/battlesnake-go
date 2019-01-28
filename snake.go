package main // import "github.com/tonobo/battlesnake-go"

type SnakePoint struct {
	Point
	Snake  *Snake `json:"-"`
	IsHead bool
	IsTail bool
}

func (sp *SnakePoint) Type() string {
	return "snake_point"
}

type Snake struct {
	ID         string `json:"id"`
	InternalID string
	Name       string        `json:"name"`
	Health     int           `json:"health"`
	Body       []*SnakePoint `json:"body"`

	Board      *Board `json:"-"`
	Me         *Snake `json:"-"`
	aimForFood *bool
}

func (s *Snake) AimForFood() bool {
	if s.aimForFood != nil {
		return *s.aimForFood
	}
	s.aimForFood = boolPtr(false)
	if len(s.Body) < SnakeMinLenth {
		s.aimForFood = boolPtr(true)
		return *s.aimForFood
	}
	if s.Health < FoodHealthLimit {
		s.aimForFood = boolPtr(true)
		return *s.aimForFood
	}
	for _, snake := range s.Board.Snakes {
		if snake.Me != nil && len(snake.Body) > len(s.Body) {
			// Another snake is larger
			s.aimForFood = boolPtr(true)
			return *s.aimForFood
		}
	}
	return *s.aimForFood
}
