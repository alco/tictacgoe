package gameloop

import (
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"tictacgoe/game"
)

// Possible state for the connection to be in
const (
	kStateMyTurn = iota
	kStateHisTurn
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

// Message format for kMessageTurn
type TurnData struct {
	Coords [2]int
	Result int
}

type Loop struct {
	*game.Board
	GameResult int
	Commands   chan int

	responseChan chan cmdStruct
	firstPlayer  int
	lastError    error
}

func NewLoop() *Loop {
	var n Loop
	n.Board = game.NewBoard()
	n.Commands = make(chan int)
	n.responseChan = make(chan cmdStruct)
	return &n
}

// Interface for the client

func (n *Loop) SendResponse(msgType int, payload interface{}) {
	n.responseChan <- cmdStruct{msgType, payload}
}

func (n *Loop) Error() error {
	return n.lastError
}

// Communicating with a client

// Sync call
func (n *Loop) callCommand(cmd int) cmdStruct {
	n.Commands <- cmd
	return <-n.responseChan
}

// Async call
func (n *Loop) castCommand(cmd int) {
	n.Commands <- cmd
}

/// Error handling

func (n *Loop) fatal(val interface{}, args ...interface{}) {
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
	panic(err) // this panic will be caught in handleConnection (unless it's a runtime error)
}

/// Establishing a connection

func genFirstPlayer() int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(2)
}

func (n *Loop) Listen(address string) error {
	n.firstPlayer = genFirstPlayer()
	go n.handleConnection()
	return nil
}

func (n *Loop) ConnectToServer(address string) error {
	n.firstPlayer = genFirstPlayer()
	go n.handleConnection()
	return nil
}

/// Common

// When the game is finished, we set our local result value for the view to be
// able to display it (since the view is not aware whether we are the first
// player or second)
func (n *Loop) checkResult(result int) bool {
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
func (n *Loop) handleConnection() {
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

			var gameFinished = n.checkResult(turn.Result)
			if gameFinished {
				n.castCommand(CmdGameFinished)
				return
			}
			state = kStateHisTurn

		case kStateHisTurn:
			n.castCommand(CmdWaitForOpponent)

			// Sleep one second to let the view display whatever it wants
			time.Sleep(time.Second / 2)

			result, err := n.MakeAIMove()
			if err != nil {
				n.fatal("Invalid move received from peer: %v", err)
			}

			var gameFinished = n.checkResult(result)
			if gameFinished {
				n.castCommand(CmdGameFinished)
				return
			}
			state = kStateMyTurn

		}
	}
}
