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
}

func worker(world [][]byte, p Params, c distributorChannels, turn int, workerOut chan<- byte, workerHeight int) {
	tempWorld := make([][]byte, p.ImageHeight+2)
	for i := range world {
		tempWorld[i] = make([]byte, p.ImageWidth)
	}

	for y := 1; y <= workerHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			numAliveNeighbours := aliveNeighbours(world, y, x, p)
			if world[y][x] != 0 {
				if numAliveNeighbours == 2 || numAliveNeighbours == 3 {
					tempWorld[y][x] = 255
				} else {
					tempWorld[y][x] = 0
					c.events <- CellFlipped{turn, util.Cell{Y: y, X: x}}
				}
			} else {
				if numAliveNeighbours == 3 {
					tempWorld[y][x] = 255
					c.events <- CellFlipped{turn, util.Cell{Y: y, X: x}}
				} else {
					tempWorld[y][x] = 0
				}
			}
		}
	}
	for y := 0; y < workerHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			workerOut <- tempWorld[y+1][x]
		}
	}
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

	currentAliveCells := make([]util.Cell, 0) // create aliveCells slice
	for y := 0; y < p.ImageHeight; y++ {      // go through all cells in world
		for x := 0; x < p.ImageWidth; x++ {
			val := <-c.ioInput
			if val != 0 {
				currentAliveCells = append(currentAliveCells, util.Cell{X: x, Y: y}) // adds current cell to the aliveCells slice
				world[y][x] = 255                                                    // update value of current cell
			}
		}
	}

	for _, cell := range currentAliveCells {
		c.events <- CellFlipped{turn, cell} // sends CellFlipped event for all alive cells
	}

	// implement for left over pixels using mod e.g. 256 not divisible by 5 threads

	for turns := 0; turns < p.Turns; turns++ {
		workerOut := make([]chan byte, p.Threads)
		workerHeight := p.ImageHeight / p.Threads
		remainderHeight := p.ImageHeight % p.Threads

		for thread := 0; thread < p.Threads; thread++ {
			var currentSplit [][]byte

			workerHeightWithRemainder := workerHeight + remainderHeight
			workerOut[thread] = make(chan byte)

			if thread == p.Threads-1 {
				currentSplit = splitWorldLastThread(world, workerHeight, workerHeightWithRemainder, p, thread)
				go worker(currentSplit, p, c, turns, workerOut[thread], workerHeightWithRemainder)
			} else {
				currentSplit = splitWorld(world, workerHeight, p, thread)
				go worker(currentSplit, p, c, turns, workerOut[thread], workerHeight)
			}

		}
		for thread := 0; thread < p.Threads; thread++ {
			var splitHeight int
			if thread == (p.Threads-1) && remainderHeight > 0 {
				splitHeight = workerHeight + remainderHeight
			} else {
				splitHeight = workerHeight
			}

			newSplit := make([][]byte, splitHeight)
			for i := range newSplit {
				newSplit[i] = make([]byte, p.ImageWidth)
			}

			for y := 0; y < splitHeight; y++ {
				for x := 0; x < p.ImageWidth; x++ {
					newSplit[y][x] = <-workerOut[thread]
				}
			}
			for y := 0; y < splitHeight; y++ {
				for x := 0; x < p.ImageWidth; x++ {
					world[thread*workerHeight+y][x] = newSplit[y][x]
				}
			}
		}

		c.events <- TurnComplete{turns}
	}

	finalAliveCells := make([]util.Cell, 0)
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == 255 {
				cell := util.Cell{Y: y, X: x}
				finalAliveCells = append(finalAliveCells, cell)
			}
		}
	}

	c.events <- FinalTurnComplete{p.Turns, finalAliveCells}

	// TODO: Execute all turns of the Game of Life.

	// TODO: Send correct Events when required, e.g. CellFlipped, TurnComplete and FinalTurnComplete.
	//		 See event.go for a list of all events.

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{p.Turns, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
