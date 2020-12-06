package gol

import(
	"bufio"
	"fmt"
	"net"
	"strconv"
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


func engineConnection(p Params, c distributorChannels) {
	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}


	c.ioCommand <- ioCommand(ioInput)                             // send read command down command channel
	filename := fmt.Sprintf("%dx%d", p.ImageHeight, p.ImageWidth) // gets file name from putting file dimensions together
	c.ioFilename <- filename                                      // sends file name to the fileName channel

	for y:=0;y<p.ImageHeight;y++{
		for x:=0; x<p.ImageWidth; x++{
			world[y][x] = <-c.ioInput
		}
	}

	
	
}