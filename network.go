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
	kStateMyTurn
	kStateHisTurn
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

type TurnData struct {
	Coords [2]int
	Result int
}

func handleConnection(conn net.Conn, cmdChan chan int, responseChan chan Cmd, state int) {
	// We expect only two possible states at first
	var firstPlayer int
	if state == kStateStartTurnNegotiation {
		firstPlayer = negotiateTurn(conn)
	} else if state == kStateWaitForTurnNegotiation {
		firstPlayer = validateTurn(conn)
	} else {
		panic(fmt.Sprintf("Unexpected state: %v", state))
	}

	if firstPlayer == 0 {
		ownChar, oppChar = 'X', 'O'
		state = kStateMyTurn
	} else {
		ownChar, oppChar = 'O', 'X'
		state = kStateHisTurn
	}

	// The game has started
	fmt.Printf("The game has started: %v\n", firstPlayer)
	for {
		switch state {
		case kStateMyTurn:
			cmdChan <- kCmdMakeTurn
			var cmd = <-responseChan
			var turn = cmd.payload.(TurnData)
			if turn.Result > GameFinished {
				panic("GAME FINISHED")
			}
			if turn.Result != OKMove {
				panic(fmt.Sprintf("Unexpected result %v", turn.Result))
			}

			sendMessage(conn, "turn", turn)
			cmdChan <- kCmdWaitForOpponent
			state = kStateHisTurn

		case kStateHisTurn:
			var turn TurnData
			expectMessage(conn, "turn", &turn)

			// Validate peer's move
			result, err := board.makeMove(turn.Coords, oppChar)
			if err != nil {
				printError(err)
				panic(err)
			}

			if result != turn.Result {
				sendMessage(conn, "fatal", "Mismatching turn result")
			}

			state = kStateMyTurn
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

func deserialize(r io.Reader, value interface{}) {
	var dec = gob.NewDecoder(r)
	err := dec.Decode(value)
	if err != nil {
		panic(err)
	}
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

func sendMessage(conn net.Conn, msg string, value interface{}) {
	fmt.Printf(">> Sending message (%v, %v)\n", msg, value)

	err := writeObj(conn, msg, value)
	if err != nil {
		panic(err)
	}
}

func receiveMessage(conn net.Conn) string {
	var byteBuf = make([]byte, 1)
	var buf []byte
	var msg string
	for {
		_, err := conn.Read(byteBuf)
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
	fmt.Println(">> Received message: ", msg)
	return msg
}

func expectMessage(conn net.Conn, expectedMsg string, value interface{}) {
	msg := receiveMessage(conn)
	if msg != expectedMsg {
		panic(fmt.Sprintf("Unexpected message %v", msg))
	}

	deserialize(conn, value)
}

func invertPlayer(player int) int {
	return 1 - player
}

func negotiateTurn(conn net.Conn) int {
	println("kStateStartTurnNegotiation")

	var timestamp = time.Now().Unix()
	sendMessage(conn, "timestamp", timestamp)

	//msg, value := receiveMessage(conn)
	var otherFirstPlayer int
	expectMessage(conn, "firstPlayer", &otherFirstPlayer)

	println("Other first player = ", otherFirstPlayer)
	rand.Seed(timestamp)
	var firstPlayer = rand.Intn(2)
	if firstPlayer != otherFirstPlayer {
		sendMessage(conn, "fatal", "Mismatching first player")
		panic("Mismatching first player")
	}

	// Confirm chosen first player
	sendMessage(conn, "firstPlayer", firstPlayer)
	return invertPlayer(firstPlayer)
}

func validateTurn(conn net.Conn) int {
	var timestamp int64
	expectMessage(conn, "timestamp", &timestamp)

	mytime := time.Now().Unix()
	if abs(mytime - timestamp) > 1 {
		println("Bad timestamp")
		sendMessage(conn, "fatal", "bad timestamp")
		panic("bad timestamp")
	}

	rand.Seed(int64(timestamp))
	var turn = rand.Intn(2)
	sendMessage(conn, "firstPlayer", turn)
	var otherTurn int
	expectMessage(conn, "firstPlayer", &otherTurn)
	if turn != otherTurn {
		sendMessage(conn, "fatal", "Mismatching first player")
		panic("Mismatching first player")
	}
	return turn
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
