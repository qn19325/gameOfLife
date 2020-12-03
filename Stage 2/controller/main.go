package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
)

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

func main() {
	// Get the server address and port from the commandline arguments
	addrPtr := flag.String("ip", "127.0.0.1:8030", "IP:port string to connect to")
	flag.Parse()

	conn, _ := net.Dial("tcp", *addrPtr)

	go read(&conn)
	write(&conn)

}
