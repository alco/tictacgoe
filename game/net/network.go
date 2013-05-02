package gamenet

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"runtime"
	"time"

	"tictacgoe/game"
)

// Possible state for the connection to be in
const (
	kStateMyTurn = iota
	kStateHisTurn
	kStateWaitForResultConfirmation
)

// Commands used for communication with the view module of the program
const (
	CmdMakeTurn = iota
	CmdWaitForOpponent
	CmdWaitForResultConfirmation
	CmdGameFinished
	CmdHandleError
)

// The final outcome of a match
const (
	GameResultDraw = iota
	GameResultMeWin
	GameResultHeWin
)

// Message format for exchange with the view module
type cmdStruct struct {
	msgType int
	payload interface{}
}

// Supported messages
const (
	kMessageTurn = "turn"
	kMessageFatal = "fatal"
	kMessageWinStatus = "winstatus"
	kMessageWinStatusConfirm = "winstatusConfirmation"
	kMessageTimestamp = "timestamp"
	kMessageFirstPlayer = "firstPlayer"
)

// Message format for kMessageTurn
type TurnData struct {
	Coords [2]int
	Result int
}

type Net struct {
	*game.Board
	GameResult int
	Commands   chan int

	conn         net.Conn
	firstPlayer  int
	responseChan chan cmdStruct
	lastError    error
}

func NewNet() *Net {
	var n Net
	n.Board = game.NewBoard()
	n.Commands = make(chan int)
	n.responseChan = make(chan cmdStruct)
	return &n
}

// Interface for the client

func (n *Net) SendResponse(msgType int, payload interface{}) {
	n.responseChan <- cmdStruct{msgType, payload}
}

func (n *Net) Error() error {
	return n.lastError
}

// Communicating with a client

// Sync call
func (n *Net) callCommand(cmd int) cmdStruct {
	n.Commands <- cmd
	return <-n.responseChan
}

// Async call
func (n *Net) castCommand(cmd int) {
	n.Commands <- cmd
}

/// Error handling

func (n *Net) fatal(val interface{}, args ...interface{}) {
	var err error
	switch val.(type) {
	case string:
		if len(args) > 0 {
			err = errors.New(fmt.Sprintf(val.(string), args...))
		} else {
			err = errors.New(val.(string))
		}
	case error:
		err = val.(error)
	}
	n.conn.Close()
	panic(err) // this panic will be caught in handleConnection (unless it's a runtime error)
}

func (n *Net) fatalSend(errMsg string) {
	n.sendMessage(kMessageFatal, errMsg)
	n.conn.Close()

	var err = errors.New(errMsg)
	panic(err) // this panic will be caught in handleConnection (unless it's a runtime error)
}

/// Establishing a connection

func (n *Net) Listen(address string) (err error) {
	ln, err := net.Listen("tcp", address)
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

// When the game is finished, we set our local result value for the view to be
// able to display it (since the view is not aware whether we are the first
// player or second)
func (n *Net) checkResult(result int) bool {
	if result > game.GameFinished {
		if result == game.Draw {
			n.GameResult = GameResultDraw
		} else if (result == game.Player1Win && n.firstPlayer == 0) ||
			(result == game.Player2Win && n.firstPlayer == 1) {
			n.GameResult = GameResultMeWin
		} else {
			n.GameResult = GameResultHeWin
		}
		return true
	}
	return false
}

// Run loop of our program. Handles communication with the other peer and with
// the view module (for querying user input and display game progress)
func (n *Net) handleConnection() {
	// Trap all panics except runtime errors
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}

			n.lastError = r.(error)
			n.castCommand(CmdHandleError)
		}
	}()

	n.Board.SetFirstPlayer(n.firstPlayer)

	var state int
	if n.firstPlayer == 0 {
		state = kStateMyTurn
	} else {
		state = kStateHisTurn
	}

	for {
		switch state {
		case kStateMyTurn:
			// Get turn data from the view module
			var cmd = n.callCommand(CmdMakeTurn)

			var turn = cmd.payload.(TurnData)
			n.sendMessage(kMessageTurn, turn)

			var gameFinished = n.checkResult(turn.Result)
			if gameFinished {
				n.castCommand(CmdWaitForResultConfirmation)
				state = kStateWaitForResultConfirmation
			} else {
				state = kStateHisTurn
			}

		case kStateHisTurn:
			n.castCommand(CmdWaitForOpponent)

			var turn TurnData
			n.expectMessage(kMessageTurn, &turn)

			// Validate peer's move
			result, err := n.MakeOppMove(turn.Coords)
			if err != nil {
				n.fatal("Invalid move received from peer: %v", err)
			}

			// Sanity check against cheating
			if result != turn.Result {
				n.fatalSend("Mismatching turn result")
			}

			var gameFinished = n.checkResult(result)
			if gameFinished {
				// Confirm game result with the peer

				n.sendMessage(kMessageWinStatus, result)

				var resultsMatch bool
				n.expectMessage(kMessageWinStatusConfirm, &resultsMatch)

				if resultsMatch {
					n.castCommand(CmdGameFinished)
					return
				} else {
					n.fatal("Could not agree on game result: our %v", result)
				}
			}
			state = kStateMyTurn

		case kStateWaitForResultConfirmation:
			var result int
			n.expectMessage(kMessageWinStatus, &result)

			if result != n.FinalResult() {
				n.sendMessage(kMessageWinStatusConfirm, false)
				n.fatal("Failed to agree on final result: our %v vs their %v", n.FinalResult(), result)
			} else {
				n.sendMessage(kMessageWinStatusConfirm, true)
				n.castCommand(CmdGameFinished)
			}
			return
		}
	}
}

func (n *Net) sendMessage(msg string, value interface{}) {
	/*fmt.Printf(">> Sending message (%v, %v)\n", msg, value)*/

	err := writeValue(n.conn, msg, value)
	if err != nil {
		n.fatal("Failed to send message: %v", err)
	}
}

func (n *Net) receiveMessage() (string, error) {
	var byteBuf = make([]byte, 1)
	var buf []byte
	var msg string
	for {
		_, err := n.conn.Read(byteBuf)
		if err != nil {
			return "", err
		}
		b := byteBuf[0]
		if b == ';' {
			msg = string(buf)
			break
		}
		buf = append(buf, b)
	}
	/*fmt.Println(">> Received message: ", msg)*/
	return msg, nil
}

func (n *Net) expectMessage(expectedMsg string, value interface{}) {
	msg, err := n.receiveMessage()
	if err != nil {
		n.fatal(err)
	}

	if msg == "fatal" {
		var errString string
		err = readValue(n.conn, &errString)
		if err != nil {
			n.fatal("Could not process fatal error from the peer: %v", errString)
		}
		n.fatal("Got fatal error from the peer: %v", errString)
	}

	if msg != expectedMsg {
		n.fatal("Unexpected message %v", msg)
	}

	err = readValue(n.conn, value)
	if err != nil {
		n.fatal("Could not read received value: %v", err)
	}
}

// Validate our timestamp with the peer, then use it as a seed value for the
// RNG
func (n *Net) negotiateTurn() int {
	var timestamp = time.Now().Unix()
	n.sendMessage(kMessageTimestamp, timestamp)

	var otherFirstPlayer int
	n.expectMessage(kMessageFirstPlayer, &otherFirstPlayer)

	rand.Seed(timestamp)
	var firstPlayer = rand.Intn(2)
	if firstPlayer != otherFirstPlayer {
		n.fatalSend("Mismatching first player")
	}

	// Confirm chosen first player
	n.sendMessage(kMessageFirstPlayer, firstPlayer)
	return invertPlayer(firstPlayer)
}

// Check that the peer's timestamp is almost the same as ours and generate the
// first player based on it
func (n *Net) validateTurn() int {
	var timestamp int64
	n.expectMessage(kMessageTimestamp, &timestamp)

	mytime := time.Now().Unix()
	if abs(mytime-timestamp) > 1 {
		n.fatalSend("bad timestamp")
	}

	rand.Seed(timestamp)
	var firstPlayer = rand.Intn(2)
	n.sendMessage(kMessageFirstPlayer, firstPlayer)

	// Confirm chosen first player with the peer
	var otherFirstPlayer int
	n.expectMessage(kMessageFirstPlayer, &otherFirstPlayer)
	if firstPlayer != otherFirstPlayer {
		n.fatalSend("Mismatching first player")
	}
	return firstPlayer
}

/// Utility functions

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// Used by the connecting peer to correspond with the other peer's first player
// choice
func invertPlayer(player int) int {
	return 1 - player
}

// Encode the value and send it to the peer
func writeValue(conn net.Conn, msg string, obj interface{}) error {
	var data, err = serialize(msg, obj)
	if err != nil {
		return err
	}
	nbytes, err := conn.Write(data)
	if err == nil && nbytes != len(data) {
		return errors.New("Couldn't write all bytes")
	}
	return nil
}

// Read the encoded value from the reader and decode it
func readValue(r io.Reader, value interface{}) error {
	var dec = gob.NewDecoder(r)
	err := dec.Decode(value)
	return err
}

// Low-level encoding function
func serialize(msg string, obj interface{}) ([]byte, error) {
	// One packet has the following format:
	// <message string>;<gob-encoded data>
	var buf = new(bytes.Buffer)
	buf.WriteString(msg)
	buf.WriteByte(';')
	var enc = gob.NewEncoder(buf)
	err := enc.Encode(obj)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
