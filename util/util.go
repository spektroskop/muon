package util

import "fmt"

type Geometry struct {
	X, Y, Width, Height int
}

func (g Geometry) String() string {
	return fmt.Sprintf("%d+%dx%dx%d",
		g.X, g.Y, g.Width, g.Height,
	)
}

type F func() error

func Do(fs []F) error {
	for _, f := range fs {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func (g Geometry) Contains(x, y int) bool {
	return g.X <= x && x < g.X+g.Width &&
		g.Y <= y && y < g.Y+g.Height
}
