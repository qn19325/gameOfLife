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

// WorkerStartHeight is a global variable for the current worker's starting height
var WorkerStartHeight int

// Worker function that calculates the new state of each cell within it's boundaries
func worker(world [][]byte, p Params, c distributorChannels, workerOut chan<- byte, workerHeight int) {
	// Create a temporary empty world
	tempWorld := createWorld(p.ImageHeight+2, p.ImageWidth)

	// Loop through the worker's section of the world
	for y := 1; y <= workerHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			// Get number of alive neighbours
			numAliveNeighbours := aliveNeighbours(world, y, x, p)

			// Calculate what's the new state of the cell depending on alive neighbours
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

	// Send the updated world down the 'workerOut' channel
	for y := 0; y < workerHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			workerOut <- tempWorld[y+1][x]
		}
	}
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// Create a 2D slice to store the world.
	world := createWorld(p.ImageHeight, p.ImageWidth)

	// Send read command down command channel
	c.ioCommand <- ioCommand(ioInput)
	// Gets file name from putting file dimensions together
	filename := fmt.Sprintf("%dx%d", p.ImageHeight, p.ImageWidth)
	// Sends file name to the fileName channel
	c.ioFilename <- filename

	// Calculate which cells are alive at the start
	currentAliveCells := make([]util.Cell, 0)
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			val := <-c.ioInput
			if val != 0 {
				// Adds current cell to the aliveCells slice
				currentAliveCells = append(currentAliveCells, util.Cell{X: x, Y: y})
				// Update value of current cell
				world[y][x] = 255
			}
		}
	}

	// For all initially alive cells send a CellFlipped Event.
	for _, cell := range currentAliveCells {
		c.events <- CellFlipped{0, cell}
	}

	// Create a new ticker that goes off every 2 seconds
	ticker := time.NewTicker(2 * time.Second)

	// Loop through all the turns for this game
	for turn := 0; turn < p.Turns; turn++ {
		select {
		// If a key is pressed
		case pressed := <-c.keyPresses:
			if pressed == 's' {
				// If 's' is pressed then save the current world as a pgm file
				outputPGM(world, c, p, turn)
			} else if pressed == 'q' {
				// If 'q' is pressed
				// then save the current world as a pgm file
				outputPGM(world, c, p, turn)

				// Quit the execution
				c.events <- StateChange{CompletedTurns: turn, NewState: Quitting}
				c.ioCommand <- ioCheckIdle
				<-c.ioIdle
				close(c.events)
				return
			} else if pressed == 'p' {
				// If 'p' is pressed, then pause the game until 'p' is pressed again
				c.events <- StateChange{CompletedTurns: turn, NewState: Paused}
				for {
					tempKey := <-c.keyPresses
					if tempKey == 'p' {
						c.events <- StateChange{CompletedTurns: turn, NewState: Executing}
						break
					}
				}
			}
		case <-ticker.C:
			// When the ticker ticks, find out the number of alive cells in the world
			aliveCellsNum := 0
			for y := 0; y < p.ImageHeight; y++ {
				for x := 0; x < p.ImageWidth; x++ {
					if world[y][x] == 255 {
						aliveCellsNum++
					}
				}
			}
			// Notify the user about the number of alive cells
			c.events <- AliveCellsCount{
				CompletedTurns: turn,
				CellsCount:     aliveCellsNum,
			}
		default:
		}

		// A channel for sending the updated world
		workerOut := make([]chan byte, p.Threads)
		// Height of each worker
		workerHeight := p.ImageHeight / p.Threads
		// Height of the remainder that isn't inluded in each worker
		remainderHeight := p.ImageHeight % p.Threads
		WorkerStartHeight = 0

		for thread := 0; thread < p.Threads; thread++ {
			workerOut[thread] = make(chan byte)

			// Spread the remainder height across the first few threads
			var splitHeight int
			if remainderHeight > 0 {
				splitHeight = workerHeight + 1
			} else {
				splitHeight = workerHeight
			}

			// Split the world up into similar sized sections and receive the current split
			currentSplit := splitWorld(world, splitHeight, thread, turn, p)
			// Send that section to the worker to be processed
			go worker(currentSplit, p, c, workerOut[thread], splitHeight)

			// Increment the WorkerStartHeight so it's ready for the next section
			WorkerStartHeight += splitHeight
			// Decrement the remainderHeight
			if remainderHeight > 0 {
				remainderHeight--
			}
		}
		// Reset the remainderHeight and WorkerStartHeight values
		remainderHeight = p.ImageHeight % p.Threads
		WorkerStartHeight = 0

		for thread := 0; thread < p.Threads; thread++ {
			var splitHeight int
			if remainderHeight > 0 {
				splitHeight = workerHeight + 1
			} else {
				splitHeight = workerHeight
			}
			// Create an empty 2D slice for this thread's worker
			newSplit := createWorld(splitHeight, p.ImageWidth)

			// Get all the updated cells from the 'workerOut' channel
			for y := 0; y < splitHeight; y++ {
				for x := 0; x < p.ImageWidth; x++ {
					// Get the new value
					newSplit[y][x] = <-workerOut[thread]

					// Update that cell in 'world'
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

		// Send an event for completing that turn
		c.events <- TurnComplete{turn}
	}

	// Calculate the final alive cells
	finalAliveCells := make([]util.Cell, 0)
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == 255 {
				cell := util.Cell{Y: y, X: x}
				finalAliveCells = append(finalAliveCells, cell)
			}
		}
	}

	// Send an event telling the SDL the new state of the board
	c.events <- FinalTurnComplete{p.Turns, finalAliveCells}
	// Output the world into a PGM image
	outputPGM(world, c, p, p.Turns)

	// Quit the execution
	c.events <- StateChange{p.Turns, Quitting}
	close(c.events)
}
