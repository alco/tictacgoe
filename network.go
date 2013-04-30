package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"os"
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
	doneChan    chan bool
	lastResult  int
}

func (b *BoardRPC) MakeMove(coords [2]int, clientResult *int) error {
	result, err := b.makeMove(coords, oppChar)
	if err != nil {
		return err
	}

	b.lastResult = result
	*clientResult = result
	return nil
}

type MoveData struct {
	Coords [2]int
	Result int
}

func (b *BoardRPC) MakeOwnMove(dummy int, data *MoveData) error {
	println("\n<<< \x1b[1mYour turn\x1b[0m >>>")

	var result int
	var coords [2]int
	for {
		b.draw()

		var move, err = getMove()
		if err != nil {
			return err
			os.Exit(1)
		}

		coords, err = parseMove(move)
		if err != nil {
			printError(err)
			continue
		}

		result, err = b.makeMove(coords, ownChar)
		if err != nil {
			printError(err)
			continue
		}

		break
	}

	b.draw()
	println()
	println("Waiting for opponent...")

	b.lastResult = result
	data.Coords = coords
	data.Result = result
	return nil
}

func (b *BoardRPC) FinishGame(clientResult int, void *int) error {
	if invert(clientResult) != b.lastResult {
		return errors.New("Mismatching results between server and client")
	}
	b.draw()
	println()
	switch b.lastResult {
	case Draw:
		fmt.Println("*** \x1b[7mIt's a draw\x1b[0m ***")
	case MeWin:
		fmt.Println("*** \x1b[42m\x1b[30mYou win!\x1b[0m ***")
	case HeWin:
		fmt.Println("*** \x1b[41m\x1b[30mYou lose!\x1b[0m ***")
	}
	b.doneChan <- true
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

/// Server API

func listen(b *Board, port int) (int, chan bool) {
	board := &BoardRPC{}
	board.Board = b
	board.comChan = make(chan int)
	board.doneChan = make(chan bool)

	err := rpc.Register(board)
	if err != nil {
		panic(err)
	}
	go accept(port)

	return <-board.comChan, board.doneChan
}

func accept(port int) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
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
