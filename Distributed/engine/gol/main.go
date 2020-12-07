package main

import (
	"flag"
	"log"
	"net"
	"net/rpc"
	"runtime"
)

type Engine struct {
}
type args struct {
	world [][]byte
	turn  int
	p     Params
	c     distributorChannels
}

// -----Run function-----
// starts the distributor function with world and number of turns
func (engine *Engine) Run(args Args, reply *[][]byte) error {
	go distributor(args.p, args.c, args.world)
}

// function to send current world to client
func (engine *Engine) getCurrentWorld()

// function to get current turn

// main is the function called when starting Game of Life with 'go run .'
func main() {
	// -----THIS MAY BE WRONG-----

	runtime.LockOSThread() // not sure what this does but was in skeleton
	// Port for connection to controller
	portPtr := flag.String("port", ":8030", "listening on this port")
	flag.Parse()                             // call after all flags are defined to parse command line into flags
	rpc.Register(&Engine)                    // WHAT DOES THIS DO?
	ln, error := net.Listen("tcp", *portPtr) // listens for connections
	if error != nil {                        // produces error message if fails to connect
		log.Fatal("Unable to connect:", error)
	}
	defer ln.Close() // stops execution until surrounding functions return
	rpc.Accept(ln)   // accepts connections on ln and serves requests to server for each incoming connection

}
