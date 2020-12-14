package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"runtime"
)

// Args COMMENT
type Args struct {
	P     Params
	World [][]byte
}

type Engine struct{}

type AliveCellsReply struct {
	AliveCells     int
	CompletedTurns int
}

type SaveReply struct {
	CompletedTurns int
	World          [][]byte
}

type PauseReply struct {
	CompletedTurns int
	World          [][]byte
}

type IsAlreadyRunningReply struct {
	P                Params
	World            [][]byte
	IsAlreadyRunning bool
}

type NodeArgs struct {
	P               Params
	World           [][]byte
	NextAddress     string
	PreviousAddress string
	WorkerHeight    int
}

var WORLD [][]byte
var PARAMS Params
var ALIVE_CELLS int
var COMPLETED_TURNS = 0
var NUMBER_OF_CONTINUES = 0

// NUMBER_OF_NODES is 1 by default and then should be increased to equal the number of nodes used
var NUMBER_OF_NODES = 1

// NODE_ADDRESSES holds a slice of all the nodes' ip addresses in this format: "ip:port"
var NODE_ADDRESSES = []string{"127.0.0.1:8031", "127.0.0.1:8032"}

var Server = make([]rpc.Client, NUMBER_OF_NODES)

var PAUSE_CHANNEL = make(chan bool, 1)
var KILL_CHANNEL = make(chan bool)
var KILL_DONE_CHANNEL = make(chan bool)
var FINISHED_CHANNEL = make(chan [][]byte, 1)
var CANCEL_CHANNEL = make(chan bool, 1)
var DONE_CANCELING_CHANNEL = make(chan bool, 1)

// IsAlreadyRunning function
func (e *Engine) IsAlreadyRunning(p Params, reply *bool) (err error) {
	if COMPLETED_TURNS-1 > 0 {
		if PARAMS == p {
			*reply = true
			return
		}
		//break the already running distributor and then reply false to set up a new one
		CANCEL_CHANNEL <- true
		<-DONE_CANCELING_CHANNEL
		*reply = false
		return
	}
	*reply = false
	return
}

// Start function
func (e *Engine) Start(args Args, reply *[][]byte) (err error) {
	PARAMS = args.P
	WORLD = args.World

	if NUMBER_OF_NODES == 1 {
		WORLD = distributor(args.P, args.World)
	} else {
		var tempWorld = make([][][]byte, NUMBER_OF_NODES)

		workerHeight := args.P.ImageHeight / NUMBER_OF_NODES
		workerStartHeight := 0

		for turn := 0; turn < PARAMS.Turns; turn++ {
			select {
			case <-CANCEL_CHANNEL:
				fmt.Println("DELETING PREVIOUS ENGINE")
				ALIVE_CELLS = 0
				COMPLETED_TURNS = 0
				for i := 0; i < NUMBER_OF_CONTINUES; i++ {
					FINISHED_CHANNEL <- WORLD
				}
				NUMBER_OF_CONTINUES = 0
				DONE_CANCELING_CHANNEL <- true
				fmt.Println("DONE RESETTING")
			case pause := <-PAUSE_CHANNEL:
				if pause == true {
					for {
						tempKey := <-PAUSE_CHANNEL
						if tempKey == false {
							break
						}
					}
				}
			case <-KILL_CHANNEL:
				for node := 0; node < NUMBER_OF_NODES; node++ {
					var res int
					Server[node].Call("Node.Kill", 0, &res)
					Server[node].Close()
				}
				KILL_DONE_CHANNEL <- true
				return
			default:
				remainderHeight := args.P.ImageHeight % NUMBER_OF_NODES
				workerStartHeight = 0
				var updatedWorldResponses = make([][][]byte, NUMBER_OF_NODES)

				for node := 0; node < NUMBER_OF_NODES; node++ {
					if turn == 0 {
						Server[node] = *nodeConnection(NODE_ADDRESSES[node])
					}

					var splitHeight int
					if remainderHeight > 0 {
						splitHeight = workerHeight + 1
					} else {
						splitHeight = workerHeight
					}

					tempWorld[node] = make([][]byte, splitHeight)
					for i := range tempWorld[node] {
						tempWorld[node][i] = make([]byte, PARAMS.ImageWidth)
					}

					for y := 0; y < splitHeight; y++ {
						for x := 0; x < PARAMS.ImageWidth; x++ {
							tempWorld[node][y][x] = WORLD[workerStartHeight+y][x]
						}
					}

					request := NodeArgs{
						P:            args.P,
						World:        tempWorld[node],
						WorkerHeight: splitHeight,
					}
					var response int
					Server[node].Call("Node.SendData", request, &response)

					remainderHeight--
					workerStartHeight += splitHeight
				}

				for node := 0; node < NUMBER_OF_NODES; node++ {
					nextAddress := (node + 1 + NUMBER_OF_NODES) % NUMBER_OF_NODES
					prevAddress := (node - 1 + NUMBER_OF_NODES) % NUMBER_OF_NODES

					request := NodeArgs{
						NextAddress:     NODE_ADDRESSES[nextAddress],
						PreviousAddress: NODE_ADDRESSES[prevAddress],
					}
					var response int
					Server[node].Call("Node.SendAddresses", request, &response)
				}

				for node := 0; node < NUMBER_OF_NODES; node++ {
					// server := nodeConnection(NODE_ADDRESSES[node])
					var response [][]byte
					Server[node].Call("Node.Start", 0, &response)
					updatedWorldResponses[node] = response
				}

				WORLD = nil
				for i := 0; i < NUMBER_OF_NODES; i++ {
					WORLD = append(WORLD, updatedWorldResponses[i]...)
				}

				ALIVE_CELLS = getNumAliveCells(PARAMS, WORLD)
				COMPLETED_TURNS = turn + 1
			}
		}
	}
	ALIVE_CELLS = 0
	COMPLETED_TURNS = 0
	for i := 0; i < NUMBER_OF_CONTINUES; i++ {
		FINISHED_CHANNEL <- WORLD
	}
	NUMBER_OF_CONTINUES = 0
	*reply = WORLD

	return
}

// Continue function
func (e *Engine) Continue(x int, reply *[][]byte) (err error) {
	NUMBER_OF_CONTINUES++
	*reply = <-FINISHED_CHANNEL
	return
}

// Kill function
func (e *Engine) Kill(x int, reply *SaveReply) (err error) {
	if NUMBER_OF_NODES > 1 {
		KILL_CHANNEL <- true
		<-KILL_DONE_CHANNEL
	}

	killReply := SaveReply{
		CompletedTurns: COMPLETED_TURNS,
		World:          WORLD,
	}

	*reply = killReply
	os.Exit(0)
	return
}

// Save function
func (e *Engine) Save(x int, reply *SaveReply) (err error) {
	saveReply := SaveReply{
		CompletedTurns: COMPLETED_TURNS,
		World:          WORLD,
	}
	*reply = saveReply

	return
}

// Pause function
func (e *Engine) Pause(x int, reply *PauseReply) (err error) {
	PAUSE_CHANNEL <- true
	pauseReply := PauseReply{
		CompletedTurns: COMPLETED_TURNS,
		World:          WORLD,
	}
	*reply = pauseReply

	return
}

// Execute function
func (e *Engine) Execute(x int, reply *PauseReply) (err error) {
	PAUSE_CHANNEL <- false
	executeReply := PauseReply{
		CompletedTurns: COMPLETED_TURNS,
		World:          WORLD,
	}
	*reply = executeReply

	return
}

// Quit funtion
func (e *Engine) Quit(x int, reply *int) (err error) {
	*reply = COMPLETED_TURNS

	return
}

// GetAliveCells ...
func (e *Engine) GetAliveCells(x int, reply *AliveCellsReply) (err error) {
	aliveCells := AliveCellsReply{
		AliveCells:     ALIVE_CELLS,
		CompletedTurns: COMPLETED_TURNS,
	}
	*reply = aliveCells

	return
}

func nodeConnection(address string) *rpc.Client {
	node, error := rpc.Dial("tcp", address)

	if error != nil {
		log.Fatal("Unable to connect", error)
	}

	return node
}

// main is the function called when starting Game of Life with 'go run .'
func main() {
	runtime.LockOSThread() // not sure what this does but was in skeleton

	// Port for connection to controller
	portPtr := flag.String("port", ":8030", "listening on this port")
	flag.Parse()                             // call after all flags are defined to parse command line into flags
	rpc.Register(&Engine{})                  // WHAT DOES THIS DO?
	ln, error := net.Listen("tcp", *portPtr) // listens for connections
	if error != nil {                        // produces error message if fails to connect
		log.Fatal("Unable to connect:", error)
	}
	defer ln.Close() // stops execution until surrounding functions return
	rpc.Accept(ln)   // accepts connections on ln and serves requests to server for each incoming connection

}
