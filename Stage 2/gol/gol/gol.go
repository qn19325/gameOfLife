package gol

import (
	"flag"
	"fmt"
	"net"
)

// Params provides the details of how to run the Game of Life and which image to load.
type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

func handleError(err error) {
	// Deal with an error event.
	if err != nil {
		// print error
		fmt.Println("error")
	}
}

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func Run(p Params, events chan<- Event, keyPresses chan rune) {

	// Port for connection to controller
	portPtr := flag.String("port", ":8030", "port to listen on")
	flag.Parse()

	ln, _ := net.Listen("tcp", *portPtr)
	// handleError(err)

	// Get connection
	connection, _ := ln.Accept()
	// handleError(err)

	ioCommand := make(chan ioCommand)
	ioIdle := make(chan bool)
	ioFileName := make(chan string)
	ioOutput := make(chan uint8)
	ioInput := make(chan uint8)
	keyPressFromController := make(chan string)

	distributorChannels := distributorChannels{
		events,
		ioCommand,
		ioIdle,
		ioFileName,
		ioOutput,
		ioInput,
		keyPresses,
		keyPressFromController,
	}
	go handleController(connection, distributorChannels)
	go distributor(p, distributorChannels, connection)

	ioChannels := ioChannels{
		command:  ioCommand,
		idle:     ioIdle,
		filename: ioFileName,
		output:   ioOutput,
		input:    ioInput,
	}
	go startIo(p, ioChannels)
}
