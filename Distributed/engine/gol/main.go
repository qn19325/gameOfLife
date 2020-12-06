




package main

import (
	"flag"
	"fmt"
	"runtime"
	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/sdl"
)

// main is the function called when starting Game of Life with 'go run .'
func main() {
	// Port for connection to controller
portPtr := flag.String("port", ":8030", "port to listen on")
flag.Parse()

ln, _ := net.Listen("tcp", *portPtr)
// handleError(err)

// Get connection
connection, _ := ln.Accept()
// handleError(err)


}


