package gol

import (
	"fmt"

	"uk.ac.bris.cs/gameoflife/util"
)

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels, world [][]byte) {

	// TODO: For all initially alive cells send a CellFlipped Event.

	turn := 0

	currentAliveCells := make([]util.Cell, 0) // create aliveCells slice
	for y := 0; y < p.ImageHeight; y++ {      // go through all cells in world
		for x := 0; x < p.ImageWidth; x++ {
			val := <-c.ioInput
			if val != 0 {
				currentAliveCells = append(currentAliveCells, util.Cell{X: x, Y: y}) // adds current cell to the aliveCells slice
				world[y][x] = 1                                                      // update value of current cell
			}
		}
	}

	tempWorld := make([][]byte, p.ImageHeight)
	for i := range world {
		tempWorld[i] = make([]byte, p.ImageWidth)
	}

	for turns := 0; turns < p.Turns; turns++ {
		select {
		case pressed := <-c.keyPresses:
			if pressed == 's' {
				outputPGM(world, c, p, turns)
			} else if pressed == 'q' {
				outputPGM(world, c, p, turns)
				c.events <- StateChange{CompletedTurns: turns, NewState: Quitting}
				c.ioCommand <- ioCheckIdle
				<-c.ioIdle
				close(c.events)
				return
			} else if pressed == 'p' {
				c.events <- StateChange{CompletedTurns: turns, NewState: Paused}
				for {
					tempKey := <-c.keyPresses
					if tempKey == 'p' {
						c.events <- StateChange{CompletedTurns: turns, NewState: Executing}
						break
					}
				}
			}
		case <-ticker.C:
			aliveCellsNum := 0
			for y := 0; y < p.ImageHeight; y++ {
				for x := 0; x < p.ImageWidth; x++ {
					if world[y][x] == 255 {
						aliveCellsNum++
					}
				}
			}

			c.events <- AliveCellsCount{
				CompletedTurns: turns,
				CellsCount:     aliveCellsNum,
			}
		default:
		}
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
		for y := 0; y < p.ImageHeight; y++ {
			for x := 0; x < p.ImageWidth; x++ {
				if world[y][x] != tempWorld[y][x] {
					world[y][x] = tempWorld[y][x]
				}
			}
		}
	}
}