package main

import (
	"fmt"
    "net"
    "net/rpc"
)

type BoardRPC struct {
	*Board
}

func (b *BoardRPC) MakeMove(coords [2]int, result *int) error {
	fmt.Println("making move at coords", coords)
	*result = 13
	return nil
}

func listen(b *Board) chan int {
	board := &BoardRPC{b}
    err := rpc.Register(board)
    if err != nil {
        panic(err)
    }

	c := make(chan int)
    go accept(c)

	return c
}

func accept(c chan int) {
    l, err := net.Listen("tcp", ":8888")
    if err != nil {
        panic(err)
    }
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		panic(err)
	}

	c <- 1

	rpc.ServeConn(conn)
	println("Finished serving conn")
}
