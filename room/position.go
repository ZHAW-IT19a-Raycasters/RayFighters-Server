package room

type Coordinate struct {
	x    float64
	y    float64
	view float64
}

func NewCoordinate() Coordinate {
	return Coordinate{
		x:    0,
		y:    0,
		view: 0,
	}
}
