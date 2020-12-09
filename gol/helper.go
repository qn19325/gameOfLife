package gol

import "fmt"

func mod(x, m int) int {
	return (x + m) % m
}

func createWorld(height, width int) [][]byte {
	world := make([][]byte, height)
	for i := range world {
		world[i] = make([]byte, width)
	}
	return world
}

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

func splitWorld(world [][]byte, splitHeight, currentThread, turn int, p Params) [][]byte {
	tempWorld := createWorld(splitHeight+2, p.ImageWidth)

	for x := 0; x < p.ImageWidth; x++ {
		previousRow := (WorkerStartHeight + p.ImageHeight - 1) % p.ImageHeight
		tempWorld[0][x] = world[previousRow][x]
	}
	for x := 0; x < p.ImageWidth; x++ {
		nextRow := (WorkerStartHeight + splitHeight + p.ImageHeight) % p.ImageHeight
		tempWorld[splitHeight+1][x] = world[nextRow][x]
	}
	for y := 1; y <= splitHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			currentRow := WorkerStartHeight + y - 1
			tempWorld[y][x] = world[currentRow][x]
		}
	}
	return tempWorld
}

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
