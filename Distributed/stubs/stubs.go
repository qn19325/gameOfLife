package stubs

type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

var controllerHandler = "Engine.Run"

type Response struct {
	world      [][]byte
	aliveCells int
}

type Request struct {
	world [][]byte
	p     Params
}
