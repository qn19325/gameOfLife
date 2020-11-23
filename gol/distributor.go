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

func splitWorld(world [][]byte, workerHeight int, p Params, currentThread int) [][]byte {
	tempWorld := make([][]byte, workerHeight+2)
	for row := range tempWorld {
		tempWorld[row] = make([]byte, p.ImageWidth)
	}

	for y := 0; y < workerHeight+2; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if y == 0 {
				previousRow := (currentThread*workerHeight + p.ImageHeight - 1) % p.ImageHeight
				tempWorld[0][x] = world[previousRow][x]
			} else if y == workerHeight+1 {
				nextRow := ((currentThread+1)*workerHeight + p.ImageHeight) % p.ImageHeight
				tempWorld[workerHeight+1][x] = world[nextRow][x]
			} else {
				currentRow := currentThread*workerHeight + y - 1
				tempWorld[y][x] = world[currentRow][x]
			}
		}
	}
	return tempWorld
}

func worker(world [][]byte, p Params, c distributorChannels, turn int, workerOut chan<- byte, workerHeight int) {
	tempWorld := make([][]byte, workerHeight+2)
	for i := range world {
		tempWorld[i] = make([]byte, p.ImageWidth)
	}

	for y := 0; y < workerHeight+2; y++ {
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

	workerOut := make(chan byte)
	workerHeight := p.ImageHeight / p.Threads
	// implement for left over pixels using mod e.g. 256 not divisible by 5 threads

	for turns := 0; turns < p.Turns; turns++ {
		newWorld := make([][]byte, p.Threads)
		for thread := 0; thread < p.Threads; thread++ {
			newSplit := make([][]byte, workerHeight)
			for i := range newSplit {
				newSplit[i] = make([]byte, p.ImageWidth)
			}
			currentSplit := splitWorld(world, workerHeight, p, thread)
			go worker(currentSplit, p, c, turns, workerOut, workerHeight)
			for y := 0; y < workerHeight; y++ {
				for x := 0; x < p.ImageWidth; x++ {
					newSplit[y][x] = <-workerOut
				}
			}
			for y := 0; y < workerHeight; y++ {
				for x := 0; x < p.ImageWidth; x++ {
					//print(tempOut[y+1][x])
					world[thread*workerHeight+y][x] = newSplit[y][x]
				}
			}
		}
		for y := 0; y < p.ImageHeight; y++ {
			for x := 0; x < p.ImageWidth; x++ {
				if world[y][x] != newWorld[y][x] {
					world[y][x] = newWorld[y][x]
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
