package gol

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strconv"
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

/* ----- DON'T NEED THIS? -----
func read(conn *net.Conn) {
	// read messages from the server and display it
	reader := bufio.NewReader(*conn)
	for {
		msg, _ := reader.ReadString('\n')
		fmt.Printf(msg)
	}
}

func write(conn *net.Conn) {
	// Continually get input from the user and send messages to the server
	stdin := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Enter `p` to pause, `q` to quit or `s` to save state:")
		keyPress, _ := stdin.ReadString('\n')
		fmt.Fprintln(*conn, keyPress)
	}
}
*/

func engineConnection(p Params, c distributorChannels) {
	// Request and set the number of turns
	read := bufio.NewReader(os.Stdin)
	fmt.Println("Set number of turns: ")
	turns, _ := read.ReadString('\n')

	p.Turns, _ = strconv.Atoi(turns)

	// create slice to store world
	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}

	c.ioCommand <- ioCommand(ioInput)                             // send read command down command channel
	filename := fmt.Sprintf("%dx%d", p.ImageHeight, p.ImageWidth) // gets file name from putting file dimensions together
	c.ioFilename <- filename                                      // sends file name to the fileName channel

	// populate world
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			world[y][x] = <-c.ioInput
		}
	}

	// connect to engine
	server := flag.String("server", "127.0.0.1:8030", "IP:port string to connect to as server")
	client, error := rpc.Dial("tcp", *server)
	if error != nil {
		log.Fatal("Unable to connect", error)
	}
	// Send specified number of turns to engine

	// Send intial world to server

	// trigger the engine to start Run method

	// create ticker to send alive cells requests to engine
	// receives number of alive cells from server
	// sends AliveCellsCount event down events channel

	// implement functionality for key pressing (in select)
	// s == request current world from engine, outputPGM
	// p == send request to pause to engine (replies current turn)
	//  StateChange event (stop)
	//  stop ticker
	// p == send request to resume execution (replies current turn)
	//  StateChange event (executing)
	//  start new ticker
	// q == request current turn
	//  StateChange event (qutting)
	//  disconnect from server (without error on server)

	// stop ticker

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	//close events channel
	close(c.events)

}
