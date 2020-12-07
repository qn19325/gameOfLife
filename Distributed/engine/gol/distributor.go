package gol

import (
	"uk.ac.bris.cs/gameoflife/util"
)

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, world [][]byte) {

	// create slice to store next state of the world
	tempWorld := make([][]byte, p.ImageHeight)
	for i := range world {
		tempWorld[i] = make([]byte, p.ImageWidth)
	}

	for turns := 0; turns < p.Turns; turns++ {
		for y := 0; y < p.ImageHeight; y++ {
			for x := 0; x < p.ImageWidth; x++ {
				numAliveNeighbours := aliveNeighbours(world, y, x, p)
				if world[y][x] != 0 {
					if numAliveNeighbours == 2 || numAliveNeighbours == 3 {
						tempWorld[y][x] = 1
					} else {
						tempWorld[y][x] = 0
						c.events <- CellFlipped{turns, util.Cell{Y: y, X: x}}
					}
				} else {
					if numAliveNeighbours == 3 {
						tempWorld[y][x] = 1
						c.events <- CellFlipped{turns, util.Cell{Y: y, X: x}}
					} else {
						tempWorld[y][x] = 0
					}
				}
			}
		}
	}
	return tempWorld
}
