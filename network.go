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

type Cmd struct {
	msgType int
	payload interface{}
}

type TurnData struct {
	Coords [2]int
	Result int
}

type Net struct {
	*Board
	GameResult int

	conn net.Conn
	firstPlayer int
	cmdChan     chan int
	responseChan chan Cmd
}

func NewNet() *Net {
	var n Net
	n.Board = NewBoard()
	n.cmdChan = make(chan int)
	n.responseChan = make(chan Cmd)
	return &n
}

// Interface for the client

func (n *Net) GetCommand() int {
	return <-n.cmdChan
}

func (n *Net) SendResponse(response Cmd) {
	n.responseChan <- response
}

// Communicating with a client

// Sync call
func (n *Net) callCommand(cmd int) Cmd {
	n.cmdChan <- cmd
	return <-n.responseChan
}

// Async call
func (n *Net) castCommand(cmd int) {
	n.cmdChan <- cmd
}

// Establishing a connection

func (n *Net) Listen(port int) (err error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return
	}

	conn, err := ln.Accept()
	if err != nil {
		return
	}

	n.conn = conn
	n.firstPlayer = n.validateTurn()
	go n.handleConnection()

	return
}

func (n *Net) ConnectToServer(address string) (err error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return
	}

	n.conn = conn
	n.firstPlayer = n.negotiateTurn()
	go n.handleConnection()

	return
}

/// Common

func (n *Net) checkResult(result int) bool {
	if result > GameFinished {
		if result == Draw {
			n.GameResult = Draw
		} else if result == Player1Win && n.firstPlayer == 0 {
			n.GameResult = MeWin
		} else if result == Player2Win && n.firstPlayer == 1 {
			n.GameResult = MeWin
		} else {
			n.GameResult = HeWin
		}
		n.finalResult = result
		return true
	}
	return false
}

func (n *Net) handleConnection() {
	var state int
	if n.firstPlayer == 0 {
		ownChar, oppChar = 'X', 'O'
		state = kStateMyTurn
	} else {
		ownChar, oppChar = 'O', 'X'
		state = kStateHisTurn
	}

	// The game has started
	for {
		switch state {
		case kStateMyTurn:
			var cmd = n.callCommand(kCmdMakeTurn)

			var turn = cmd.payload.(TurnData)
			if turn.Result <= GameFinished && turn.Result != OKMove {
				panic(fmt.Sprintf("Unexpected result %v", turn.Result))
			}

			n.sendMessage("turn", turn)

			var gameFinished = n.checkResult(turn.Result)
			if gameFinished {
				n.castCommand(kCmdWaitForResultConfirmation)
				state = kStateWaitForResultConfirmation
			} else {
				state = kStateHisTurn
			}

		case kStateHisTurn:
			n.castCommand(kCmdWaitForOpponent)

			var turn TurnData
			n.expectMessage("turn", &turn)

			// Validate peer's move
			result, err := n.Board.makeMove(turn.Coords, oppChar)
			if err != nil {
				printError(err)
				panic(err)
			}

			if result != turn.Result {
				n.sendMessage("fatal", "Mismatching turn result")
			}

			var gameFinished = n.checkResult(result)
			if gameFinished {
				// Confirm game result with the peer

				n.sendMessage("winstatus", result)

				var resultsMatch bool
				n.expectMessage("winstatusConfirmation", &resultsMatch)

				if resultsMatch {
					n.castCommand(kCmdGameFinished)
					return
				} else {
					panic("Could not agree on game result")
				}
			}

			state = kStateMyTurn

		case kStateWaitForResultConfirmation:
			var result int
			n.expectMessage("winstatus", &result)

			if result != n.finalResult {
				n.sendMessage("winstatusConfirmation", false)
				panic("Failed to agree on final result")
			} else {
				n.sendMessage("winstatusConfirmation", true)
				n.castCommand(kCmdGameFinished)
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
		/*fmt.Println("Written btyes: ", nbytes)*/
	}
	return
}

func (n *Net) sendMessage(msg string, value interface{}) {
	/*fmt.Printf(">> Sending message (%v, %v)\n", msg, value)*/

	err := writeObj(n.conn, msg, value)
	if err != nil {
		panic(err)
	}
}

func (n *Net) receiveMessage() string {
	var byteBuf = make([]byte, 1)
	var buf []byte
	var msg string
	for {
		_, err := n.conn.Read(byteBuf)
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
	/*fmt.Println(">> Received message: ", msg)*/
	return msg
}

func (n *Net) expectMessage(expectedMsg string, value interface{}) {
	msg := n.receiveMessage()
	if msg != expectedMsg {
		panic(fmt.Sprintf("Unexpected message %v", msg))
	}

	deserialize(n.conn, value)
}

func invertPlayer(player int) int {
	return 1 - player
}

func (n *Net) negotiateTurn() int {
	var timestamp = time.Now().Unix()
	n.sendMessage("timestamp", timestamp)

	//msg, value := receiveMessage(conn)
	var otherFirstPlayer int
	n.expectMessage("firstPlayer", &otherFirstPlayer)

	rand.Seed(timestamp)
	var firstPlayer = rand.Intn(2)
	if firstPlayer != otherFirstPlayer {
		n.sendMessage("fatal", "Mismatching first player")
		panic("Mismatching first player")
	}

	// Confirm chosen first player
	n.sendMessage("firstPlayer", firstPlayer)
	return invertPlayer(firstPlayer)
}

func (n *Net) validateTurn() int {
	var timestamp int64
	n.expectMessage("timestamp", &timestamp)

	mytime := time.Now().Unix()
	if abs(mytime - timestamp) > 1 {
		println("Bad timestamp")
		n.sendMessage("fatal", "bad timestamp")
		panic("bad timestamp")
	}

	rand.Seed(int64(timestamp))
	var turn = rand.Intn(2)
	n.sendMessage("firstPlayer", turn)
	var otherTurn int
	n.expectMessage("firstPlayer", &otherTurn)
	if turn != otherTurn {
		n.sendMessage("fatal", "Mismatching first player")
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
