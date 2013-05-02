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

const (
	kStateMyTurn = iota
	kStateHisTurn
	kStateWaitForResultConfirmation
)

const (
	kCmdMakeTurn = iota
	kCmdWaitForOpponent
	kCmdWaitForResultConfirmation
	kCmdGameFinished
)

type BoardNet struct {
	*Board
	firstPlayer int
	comChan     chan int
}

// Server
func listen(b *Board, port int) (cmdChan chan int, responseChan chan Cmd, err error) {
	board := &BoardNet{}
	board.Board = b
	board.comChan = make(chan int)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return
	}

	conn, err := ln.Accept()
	if err != nil {
		return
	}

	var firstPlayer = validateTurn(conn)

	cmdChan = make(chan int)
	responseChan = make(chan Cmd)
	go handleConnection(conn, cmdChan, responseChan, firstPlayer)

	return
}

// Client
func connectToServer(address string) (cmdChan chan int, responseChan chan Cmd, err error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return
	}

	var firstPlayer = negotiateTurn(conn)

	cmdChan = make(chan int)
	responseChan = make(chan Cmd)
	go handleConnection(conn, cmdChan, responseChan, firstPlayer)

	return
}

/// Common

type TurnData struct {
	Coords [2]int
	Result int
}

func checkResult(result int, firstPlayer int) bool {
	if result > GameFinished {
		if result == Draw {
			board.gameResult = Draw
		} else if result == Player1Win && firstPlayer == 0 {
			board.gameResult = MeWin
		} else if result == Player2Win && firstPlayer == 1 {
			board.gameResult = MeWin
		} else {
			board.gameResult = HeWin
		}
		board.finalResult = result
		return true
	}
	return false
}

func handleConnection(conn net.Conn, cmdChan chan int, responseChan chan Cmd, firstPlayer int) {
	var state int
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
			if turn.Result <= GameFinished && turn.Result != OKMove {
				panic(fmt.Sprintf("Unexpected result %v", turn.Result))
			}

			sendMessage(conn, "turn", turn)

			var gameFinished = checkResult(turn.Result, firstPlayer)
			if gameFinished {
				cmdChan <- kCmdWaitForResultConfirmation
				state = kStateWaitForResultConfirmation
			} else {
				cmdChan <- kCmdWaitForOpponent
				state = kStateHisTurn
			}

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

			var gameFinished = checkResult(result, firstPlayer)
			if gameFinished {
				// Confirm game result with the peer

				/*cmdChan <- kCmdWaitForResultConfirmation*/
				sendMessage(conn, "winstatus", result)

				var resultsMatch bool
				expectMessage(conn, "winstatusConfirmation", &resultsMatch)

				if resultsMatch {
					cmdChan <- kCmdGameFinished
					return
				} else {
					panic("Could not agree on game result")
				}
			}

			state = kStateMyTurn

		case kStateWaitForResultConfirmation:
			var result int
			expectMessage(conn, "winstatus", &result)

			if result != board.finalResult {
				sendMessage(conn, "winstatusConfirmation", false)
				panic("Failed to agree on final result")
			} else {
				sendMessage(conn, "winstatusConfirmation", true)
				cmdChan <- kCmdGameFinished
			}
			return
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
