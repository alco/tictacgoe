package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
)

const RANDOM_TRIES = 10

func genNums() []int {
	nums := make([]int, RANDOM_TRIES)
	for i := range nums {
		nums[i] = 1 + rand.Int()%2 // 1 - my turn; 2 - his turn (server's point of view)
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

type BoardRPC struct {
	*Board
	firstPlayer int
	comChan     chan int
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

/// Server

func listen(b *Board) int {
	board := &BoardRPC{}
	board.Board = b
	board.comChan = make(chan int)

	err := rpc.Register(board)
	if err != nil {
		panic(err)
	}
	go accept()

	return <-board.comChan
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

/// Client

func connectToServer(address string) int {
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		panic(err)
	}

	var nums = genNums()
	var serverNums []int
	err = client.Call("BoardRPC.WhoseTurn", nums, &serverNums)
	if err != nil {
		panic(err)
	}

	var firstPlayer = getFirstPlayer(nums, serverNums)
	var serverOK bool
	err = client.Call("BoardRPC.AgreeOnTurn", firstPlayer, &serverOK)
	if err != nil {
		panic(err)
	}

	if !serverOK {
		panic("Could not agree on turn")
	}

	return 3 - firstPlayer // First player is from the server's point of view
}
