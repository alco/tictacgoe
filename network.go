package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"
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

func serialize(msg string, obj interface{}) []byte {
	var buf = new(bytes.Buffer)
	buf.WriteString(msg)
	buf.WriteByte(';')
	var enc = gob.NewEncoder(buf)
	err := enc.Encode(obj)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func deserialize(r io.Reader) interface{} {
	var byteBuf = make([]byte, 1)
	var buf []byte
	var msg string
	for {
		_, err := r.Read(byteBuf)
		if err != nil {
			panic(err)
		}
		b := byteBuf[0]
		if b == ';' {
			msg = string(buf)
			break
		}
		buf = append(buf, b)
	}
	fmt.Println("Got message: ", msg)

	var dec = gob.NewDecoder(r)
	switch msg {
	case "timestamp":
		var val int64
		err := dec.Decode(&val)
		if err != nil {
			panic(err)
		}
		return val

	case "firstPlayer":
		var val int
		err := dec.Decode(&val)
		if err != nil {
			panic(err)
		}
		return val
	}

	return nil
}

func writeObj(conn net.Conn, msg string, obj interface{}) (err error) {
	var data = serialize(msg, obj)
	nbytes, err := conn.Write(data)
	if err == nil && nbytes != len(data) {
		err = errors.New("Couldn't write all bytes")
	} else {
		fmt.Println("Written btyes: ", nbytes)
	}
	return
}

func readObj(conn net.Conn) (obj interface{}, err error) {
	return deserialize(conn), nil
}

func negotiateTurn(conn net.Conn, state int) (int, int) {
	for {
		switch state {
		case kStateStartTurnNegotiation:
			println("kStateStartTurnNegotiation")
			var timestamp = time.Now().Unix()
			println("timestamp =", timestamp)
			err := writeObj(conn, "timestamp", timestamp)
			if err != nil {
				panic(err)
			}

			data, err := readObj(conn)
			if err != nil {
				panic(err)
			}

			if firstPlayer, ok := data.(int); ok {
				println("firstPlayer:", firstPlayer)
			} else {
				panic(fmt.Sprintf("Received invalid data: %v %T", data, data))
			}
			state = kStateWaitForTurnNegotiation

		case kStateWaitForTurnNegotiation:
			println("kStateWaitForTurnNegotiation")
			data, err := readObj(conn)
			if err != nil {
				panic(err)
			}
			println("Received data", data)

			if timestamp, ok := data.(int64); ok {
				mytime := time.Now().Unix()
				if abs(mytime - timestamp) < 1 {
					println("Good timestamp")
					// good value
					rand.Seed(int64(timestamp))
					var turn = rand.Intn(2)
					writeObj(conn, "firstPlayer", turn)
				} else {
					println("Bad timestamp")
					writeObj(conn, "fatal", "bad timestamp")
				}
			} else {
				panic("Received invalid timestamp")
			}
		}
	}
	return 0, 0
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
