package main

import (
	"bufio"
	"errors"
	"math/rand"
	"net"
)

const RANDOM_TRIES = 10

const (
	kStateWaitForTurnNegotiation = iota
	kStateStartTurnNegotiation
)

const (
	kCmdMakeTurn = iota
	kCmdWaitForOpponent
	kCmdGameFinished
)

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

type BoardNet struct {
	*Board
	firstPlayer int
	comChan     chan int
}

/*func (b *BoardRPC) WhoseTurn(clientNums []int, result *[]int) error {*/
	/*nums := genNums()*/
	/*b.firstPlayer = getFirstPlayer(nums, clientNums)*/
	/**result = nums*/

	/*if b.firstPlayer == 0 {*/
		/*panic("Not enough numbers :(")*/
	/*}*/

	/*return nil*/
/*}*/

/*func (b *BoardRPC) AgreeOnTurn(clientTurn int, result *bool) error {*/
	/*if clientTurn == b.firstPlayer {*/
		/**result = true*/
		/*b.comChan <- b.firstPlayer*/
	/*} else {*/
		/*return errors.New("server: Could not agree on turn")*/
	/*}*/
	/*return nil*/
/*}*/

// Server
func listen(b *Board) (cmdChan chan int, responseChan chan Cmd, err error) {
	board := &BoardNet{}
	board.Board = b
	board.comChan = make(chan int)

	ln, err := net.Listen("tcp", ":8888")
	if err != nil {
		return
	}

	conn, err := ln.Accept()
	if err != nil {
		return
	}

	cmdChan = make(chan int)
	responseChan = make(chan Cmd)
	go handleConnection(conn, cmdChan, responseChan, kStateWaitForTurnNegotiation)

	return
}

// Client
func connectToServer(address string) (cmdChan chan int, responseChan chan Cmd, err error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return
	}

	cmdChan = make(chan int)
	responseChan = make(chan Cmd)
	go handleConnection(conn, cmdChan, responseChan, kStateStartTurnNegotiation)

	return

	/*var nums = genNums()*/
	/*var serverNums []int*/
	/*err = client.Call("BoardRPC.WhoseTurn", nums, &serverNums)*/
	/*if err != nil {*/
		/*panic(err)*/
	/*}*/

	/*var firstPlayer = getFirstPlayer(nums, serverNums)*/
	/*var serverOK bool*/
	/*err = client.Call("BoardRPC.AgreeOnTurn", firstPlayer, &serverOK)*/
	/*if err != nil {*/
		/*panic(err)*/
	/*}*/

	/*if !serverOK {*/
		/*panic("Could not agree on turn")*/
	/*}*/

	/*return 3 - firstPlayer // First player is from the server's point of view*/
}

/// Common

func handleConnection(conn net.Conn, cmdChan chan int, responseChan chan Cmd, state int) {
	// We expect only two possible states at first
	firstPlayer, newState := negotiateTurn(conn, state)
	println(firstPlayer)

	// The game has started
	for {
		switch newState {
		}
	}
}

func serialize(obj interface{}) []byte {
	var data = make([]byte, 0)
	return data
}

func deserialize(str string) interface{} {
	var obj = struct{}{}
	return obj
}

func writeObj(conn net.Conn, obj interface{}) (err error) {
	var data = serialize(obj)
	nbytes, err := conn.Write(data)
	if err == nil && nbytes != len(data) {
		err = errors.New("Couldn't write all bytes")
	}
	return
}

func readObj(conn net.Conn) (obj interface{}, err error) {
	var rd = bufio.NewReader(conn)
	str, err := rd.ReadString('\n')
	if err == nil {
		obj = deserialize(str)
	}
	return
}

func negotiateTurn(conn net.Conn, state int) (int, int) {
	for {
		switch state {
		case kStateStartTurnNegotiation:
			var nums = genNums()
			err := writeObj(conn, nums)
			if err != nil {
				panic(err)
			}

			data, err := readObj(conn)
			if err != nil {
				panic(err)
			}

			if serverNums, ok := data.([]int); ok {
				println(serverNums)
			} else {
				panic("Received invalid data")
			}
			state = kStateWaitForTurnNegotiation
		case kStateWaitForTurnNegotiation:
			// waiting
		}
	}
	return 0, 0
}
