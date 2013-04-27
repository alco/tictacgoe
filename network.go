package main

import (
	"errors"
	"fmt"
	"math/rand"
    "net"
    "net/rpc"
)

const RANDOM_TRIES = 10

type BoardRPC struct {
	*Board
	firstPlayer int
	comChan chan int
}

type ArbitrageData struct {
	Nums   []int
	Player int
	Index  int
}

func genNums() []int {
	nums := make([]int, RANDOM_TRIES)
	for i := range nums {
		nums[i] = 1 + rand.Int() % 2  // 1 - my turn; 2 - his turn (server's point of view)
	}
	return nums
}

func getFirstPlayer(myNums []int, hisNums []int) int {
	for i := range myNums {
		var myNum, hisNum = myNums[i], hisNums[i]
		if myNum == hisNum {
			return myNum
		}
	}
	return 0
}

func (b *BoardRPC) MakeMove(coords [2]int, result *int) error {
	fmt.Println("making move at coords", coords)
	*result = 13
	return nil
}

func (b *BoardRPC) WhoseTurn(clientNums []int, result *[]int) error {
	nums := genNums()
	b.firstPlayer = getFirstPlayer(nums, clientNums)
	*result = nums

	if b.firstPlayer == 0 {
		panic("Not enough numbers :(")
	}

	return nil
}

func (b *BoardRPC) AgreeOnTurn(clientTurn int, result *bool) error {
	if clientTurn == b.firstPlayer {
		*result = true
		b.comChan <- b.firstPlayer
	} else {
		return errors.New("server: Could not agree on turn")
	}
	return nil
}

func listen(b *Board) chan int {
	board := &BoardRPC{}
	board.Board = b
	board.comChan = make(chan int)
    err := rpc.Register(board)
    if err != nil {
        panic(err)
    }

    go accept()

	return board.comChan
}

func accept() {
    l, err := net.Listen("tcp", ":8888")
    if err != nil {
        panic(err)
    }
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		panic(err)
	}

	rpc.ServeConn(conn)
	println("Finished serving conn")
}

// Decide whose turn is first
func pregameArbitrage(client net.Conn) {
	// Generate a bunch of randomly selected turns and find the first one that
	// agrees with the other side
	nums := make([]int, RANDOM_TRIES)
	for i := range nums {
		nums[i] = rand.Int() % 2  // 0 - my turn; 1 - his turn
	}
}
