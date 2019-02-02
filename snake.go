package main // import "github.com/tonobo/battlesnake-go"

type SnakePoint struct {
	Point
	Snake       *Snake `json:"-"`
	IsHead      bool
	IsTail      bool
	EvictOnStep int
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

func (s *Snake) Enemy() bool {
	if s.Me != nil {
		return true
	}
	return false
}
