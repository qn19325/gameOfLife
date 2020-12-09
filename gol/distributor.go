package gol

import (
	"fmt"
	"time"

	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
	keyPresses <-chan rune
}

var WorkerStartHeight int

func worker(world [][]byte, p Params, c distributorChannels, turn int, workerOut chan<- byte, workerHeight int) {
	tempWorld := createWorld(p.ImageHeight+2, p.ImageWidth)

	for y := 1; y <= workerHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			numAliveNeighbours := aliveNeighbours(world, y, x, p)
			if world[y][x] == 255 {
				if numAliveNeighbours == 2 || numAliveNeighbours == 3 {
					tempWorld[y][x] = 255
				} else {
					tempWorld[y][x] = 0
				}
			} else {
				if numAliveNeighbours == 3 {
					tempWorld[y][x] = 255
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
	world := createWorld(p.ImageHeight, p.ImageWidth)

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

	ticker := time.NewTicker(2 * time.Second)

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

		workerOut := make([]chan byte, p.Threads)
		workerHeight := p.ImageHeight / p.Threads
		remainderHeight := p.ImageHeight % p.Threads
		WorkerStartHeight = 0

		for thread := 0; thread < p.Threads; thread++ {
			var currentSplit [][]byte
			workerOut[thread] = make(chan byte)

			var splitHeight int
			if remainderHeight > 0 {
				splitHeight = workerHeight + 1
			} else {
				splitHeight = workerHeight
			}

			currentSplit = splitWorld(world, splitHeight, thread, turns, p)
			go worker(currentSplit, p, c, turns, workerOut[thread], splitHeight)

			WorkerStartHeight += splitHeight
			remainderHeight--
		}
		remainderHeight = p.ImageHeight % p.Threads
		WorkerStartHeight = 0
		for thread := 0; thread < p.Threads; thread++ {
			var splitHeight int
			if remainderHeight > 0 {
				splitHeight = workerHeight + 1
			} else {
				splitHeight = workerHeight
			}
			newSplit := createWorld(splitHeight, p.ImageWidth)

			for y := 0; y < splitHeight; y++ {
				for x := 0; x < p.ImageWidth; x++ {
					newSplit[y][x] = <-workerOut[thread]
				}
			}
			for y := 0; y < splitHeight; y++ {
				for x := 0; x < p.ImageWidth; x++ {
					worldY := WorkerStartHeight + y

					if world[worldY][x] != newSplit[y][x] {
						world[worldY][x] = newSplit[y][x]
						c.events <- CellFlipped{turn, util.Cell{Y: worldY, X: x}}
					}
				}
			}
			remainderHeight--
			WorkerStartHeight += splitHeight
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

	outputPGM(world, c, p, p.Turns)

	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{p.Turns, Quitting}
	close(c.events)
}
