package gol

func mod(x, m int) int {
	return (x + m) % m
}

func aliveNeighbours(world [][]byte, x, y int, p Params) int {
	neighbours := 0
	for i := -1; i < 2; i++ {
		for j := -1; j < 2; j++ {
			if i != 0 || j != 0 {
				if world[mod(y+i, p.ImageHeight)][mod(x+j, p.ImageHeight)] != 0 {
					neighbours++
				}

			}
		}
	}
	return neighbours
}
