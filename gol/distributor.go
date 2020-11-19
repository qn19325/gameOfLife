package gol

import (
	"fmt"

	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
	aliveCells chan []util.Cell
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// TODO: Create a 2D slice to store the world.
	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}

	// TODO: For all initially alive cells send a CellFlipped Event.

	c.ioCommand <- ioCommand(ioInput)                             // send read command down command channel
	filename := fmt.Sprintf("%dx%d", p.ImageHeight, p.ImageWidth) // gets file name from putting file dimensions together
	c.ioFilename <- filename                                      // sends file name to the fileName channel

	turn := 0
	aliveCells := make([]util.Cell, 0)   // create aliveCells slice
	for y := 0; y < p.ImageHeight; y++ { // go through all cells in world
		for x := 0; x < p.ImageWidth; x++ {
			// need way to see if current cell is alive
			val := <-c.ioInput
			if val != 0 {
				aliveCells = append(aliveCells, util.Cell{X: x, Y: y}) // adds current cell to the aliveCells slice
				world[y][x] = val                                      // update value of current cell
			}
		}
	}

	for _, cell := range aliveCells {
		c.events <- CellFlipped{turn, cell} // sends CellFlipped event for all alive cells
	}

	// TODO: Execute all turns of the Game of Life.
	tempWorld := make([][]byte, p.ImageHeight)
	for i := range tempWorld {
		tempWorld[i] = make([]byte, p.ImageWidth)
	}

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			neighbours := aliveNeighbours(world, x, y, p)

			if world[y][x] == 1 {
				if neighbours == 2 || neighbours == 3 {
					tempWorld[y][x] = 1
				} else {
					tempWorld[y][x] = 0
				}
			} else {
				if neighbours == 3 {
					tempWorld[y][x] = 1
				} else {
					tempWorld[y][x] = 0
				}
			}
		}
	}

	// TODO: Send correct Events when required, e.g. CellFlipped, TurnComplete and FinalTurnComplete.
	//		 See event.go for a list of all events.

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
