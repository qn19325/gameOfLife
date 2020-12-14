package gol

import "fmt"

func mod(x, m int) int {
	return (x + m) % m
}

// Creates a 2D slice of the world depending on inputted height and width
func createWorld(height, width int) [][]byte {
	world := make([][]byte, height)
	for i := range world {
		world[i] = make([]byte, width)
	}
	return world
}

// Calculate the number of alive neighbours around the cell
func aliveNeighbours(world [][]byte, y, x int, p Params) int {
	neighbours := 0
	for i := -1; i < 2; i++ {
		for j := -1; j < 2; j++ {
			if i != 0 || j != 0 {
				if world[mod(y+i, p.ImageHeight)][mod(x+j, p.ImageWidth)] != 0 {
					neighbours++
				}

			}
		}
	}
	return neighbours
}

// Split the world into the dimensions specified
func splitWorld(world [][]byte, splitHeight, currentThread, turn int, p Params) [][]byte {
	tempWorld := createWorld(splitHeight+2, p.ImageWidth)

	// Set the previous row in tempWorld
	for x := 0; x < p.ImageWidth; x++ {
		previousRow := (WorkerStartHeight + p.ImageHeight - 1) % p.ImageHeight
		tempWorld[0][x] = world[previousRow][x]
	}

	// Set the next row in tempWorld
	for x := 0; x < p.ImageWidth; x++ {
		nextRow := (WorkerStartHeight + splitHeight + p.ImageHeight) % p.ImageHeight
		tempWorld[splitHeight+1][x] = world[nextRow][x]
	}

	// Set all the middle rows in tempWorld
	for y := 1; y <= splitHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			currentRow := WorkerStartHeight + y - 1
			tempWorld[y][x] = world[currentRow][x]
		}
	}
	return tempWorld
}

// Save the current state of the world as a PGM file
func outputPGM(world [][]byte, c distributorChannels, p Params, turn int) {
	c.ioCommand <- ioCommand(ioOutput)
	outputFileName := fmt.Sprintf("%dx%dx%d", p.ImageHeight, p.ImageWidth, turn)
	c.ioFilename <- outputFileName

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			c.ioOutput <- world[y][x]
		}
	}

	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- ImageOutputComplete{turn, outputFileName}
}
