package main

import (
	"fmt"
)

type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

func mod(x, m int) int {
	return (x + m) % m
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

func getNumAliveCells(p Params, world [][]byte) int {
	aliveCellsNum := 0
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == 255 {
				aliveCellsNum++
			}
		}
	}
	return aliveCellsNum
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, world [][]byte) [][]byte {
	PARAMS = p
	tempWorld := make([][]byte, p.ImageHeight)
	for i := range tempWorld {
		tempWorld[i] = make([]byte, p.ImageWidth)
	}

	for turns := 0; turns < p.Turns; turns++ {
		select {
		case <-CANCEL_CHANNEL:
			fmt.Println("DELETING PREVIOUS ENGINE")
			ALIVE_CELLS = 0
			COMPLETED_TURNS = 0
			for i := 0; i < NUMBER_OF_CONTINUES; i++ {
				FINISHED_CHANNEL <- world
			}
			NUMBER_OF_CONTINUES = 0
			DONE_CANCELING_CHANNEL <- true
			fmt.Println("DONE RESETTING")
			return world
		case pause := <-PAUSE_CHANNEL:
			if pause == true {
				for {
					tempKey := <-PAUSE_CHANNEL
					if tempKey == false {
						break
					}
				}
			}
		default:
			for y := 0; y < p.ImageHeight; y++ {
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
			for y := 0; y < p.ImageHeight; y++ {
				for x := 0; x < p.ImageWidth; x++ {
					world[y][x] = tempWorld[y][x]
				}
			}
			WORLD = world
			ALIVE_CELLS = getNumAliveCells(p, world)
			COMPLETED_TURNS = turns + 1
		}
	}

	ALIVE_CELLS = 0
	COMPLETED_TURNS = 0
	for i := 0; i < NUMBER_OF_CONTINUES; i++ {
		FINISHED_CHANNEL <- world
	}
	NUMBER_OF_CONTINUES = 0
	return world
}
