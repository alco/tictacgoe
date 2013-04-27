package main

import (
	"errors"
	"fmt"
	"math/rand"
    "net"
    "net/rpc"
)

type BoardRPC struct {
	*Board
	firstPlayer int
}

type ArbitrageData struct {
	Nums   []int
	Player int
	Index  int
}

func (b *BoardRPC) MakeMove(coords [2]int, result *int) error {
	fmt.Println("making move at coords", coords)
	*result = 13
	return nil
}

func (b *BoardRPC) WhoseTurn(clientNums []int, result *ArbitrageData) error {
	nums := make([]int, RANDOM_TRIES)
	for i := range nums {
		nums[i] = 1 + rand.Int() % 2  // 1 - my turn; 2 - his turn
	}
	var index = 0
	for i := range nums {
		var myNum, hisNum = nums[i], clientNums[i]
		if myNum == hisNum {
			b.firstPlayer = myNum
			index = i
			break
		}
	}
	if b.firstPlayer == 0 {
		panic("Not enough numbers :(")
	}
	fmt.Printf("Nums = %v\n", nums)
	fmt.Printf("player = %v\n", b.firstPlayer)
	fmt.Printf("index = %v\n", index)
	var adata = ArbitrageData{nums, b.firstPlayer, index}
	*result = adata
	return nil
}

func (b *BoardRPC) AgreeOnTurn(clientTurn int, result *bool) error {
	if clientTurn == b.firstPlayer {
		*result = true
	} else {
		return errors.New("server: Could not agree on turn")
	}
	return nil
}

func listen(b *Board) chan net.Conn {
	board := &BoardRPC{}
	board.Board = b
    err := rpc.Register(board)
    if err != nil {
        panic(err)
    }

	c := make(chan net.Conn)
    go accept(c)

	return c
}

func accept(c chan net.Conn) {
    l, err := net.Listen("tcp", ":8888")
    if err != nil {
        panic(err)
    }
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		panic(err)
	}

	c <- conn

	rpc.ServeConn(conn)
	println("Finished serving conn")
}

const RANDOM_TRIES = 10

// Decide whose turn is first
func pregameArbitrage(client net.Conn) {
	// Generate a bunch of randomly selected turns and find the first one that
	// agrees with the other side
	nums := make([]int, RANDOM_TRIES)
	for i := range nums {
		nums[i] = rand.Int() % 2  // 0 - my turn; 1 - his turn
	}
}
