package main

import (
	"flag"
	"net"
	"runtime"
)

// -----Run function-----
// starts the distributor function with world and number of turns

// function to get the world

// function to get number of turns specified in controller

// main is the function called when starting Game of Life with 'go run .'
func main() {
	// -----THINK NEED USE RPC IN HERE-----

	runtime.LockOSThread()
	// Port for connection to controller
	portPtr := flag.String("port", ":8030", "port to listen on")
	flag.Parse()

	ln, _ := net.Listen("tcp", *portPtr)
	// handleError(err)

	// Get connection
	connection, _ := ln.Accept()
	// handleError(err)

}
