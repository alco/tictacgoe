package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

// Wrap os.Stdin to conveniently read strings
var stdin = bufio.NewReader(os.Stdin)

// Read input from the user
func getMove() (str string, err error) {
	print("> ") // prompt

	str, err = stdin.ReadString('\n')
	if err != nil {
		return
	}
	return str[:len(str)-1], err
}

// Map ASCII letter to an integer in range 0..2
func parseLetter(char byte, coord *int) bool {
	if char >= 'a' && char <= 'c' {
		*coord = int(char - 'a')
		return true
	}
	return false
}

// Map ASCII digit to an integer in range 0..2
func parseDigit(char byte, coord *int) bool {
	if char >= '1' && char <= '3' {
		*coord = int(char - '1')
		return true
	}
	return false
}

// Convert a string representing a move to integer coordinates on the board
func parseMove(move string) (coords [2]int, err error) {
	err = errors.New("err: Invalid move.")

	move = strings.Replace(move, " ", "", -1)
	if len(move) == 0 {
		err = errors.New("err: Please make a move.")
		return
	} else if len(move) != 2 {
		return
	}

	// a1 format
	if !(parseLetter(move[0], &coords[0]) || parseLetter(move[1], &coords[0])) {
		return
	}

	// 1a format
	if !(parseDigit(move[0], &coords[1]) || parseDigit(move[1], &coords[1])) {
		return
	}

	return coords, nil
}

// Print the error in red text
func printError(err interface{}) {
	fmt.Printf("\x1b[31m%v\x1b[0m\n", err)
}


type Cmd struct {
	msgType int
	payload interface{}
}

var board *Board

// CLI flags
var mode = flag.String("mode", "server", "Which mode to run in: server or client")
var addr = flag.String("addr", "localhost", "Address to connect to")
var port = flag.Int("port", 8888, "Port to listen on or connect to")

func main() {
	flag.Parse()

	var serverMode bool
	switch *mode {
	case "server":
		serverMode = true
	case "client":
		serverMode = false
	default:
		printError("Unsupported mode")
		os.Exit(1)
	}

	fmt.Println("*** Welcome to Tic-Tac-Goe ***")
	board = NewBoard()

	var cmdChan chan int
	var responseChan chan Cmd
	var err error
	if serverMode {
		fmt.Printf("Listening on port %v...\n", *port)
		cmdChan, responseChan, err = listen(board, *port)
		if err != nil {
			printError(err)
			os.Exit(1)
		}
	} else {
		println("Connecting to server...")
		cmdChan, responseChan, err = connectToServer(fmt.Sprintf("%v:%v", *addr, *port))
		if err != nil {
			printError(err)
			os.Exit(1)
		}
	}

	/*var myTurn bool*/
	/*if firstPlayer == 1 {*/
		/*ownChar, oppChar = 'X', 'O'*/
		/*myTurn = true*/
	/*} else {*/
		/*ownChar, oppChar = 'O', 'X'*/
		/*myTurn = false*/
	/*}*/

	for {
		switch <-cmdChan {
		case kCmdMakeTurn:
			println("\n<<< \x1b[1mYour turn\x1b[0m >>>")

			for {
				board.draw()

				var move, err = getMove()
				if err != nil {
					os.Exit(0)
				}

				coords, err := parseMove(move)
				if err != nil {
					printError(err)
					continue
				}

				result, err := board.makeMove(coords, ownChar)
				if err != nil {
					printError(err)
					continue
				}

				responseChan <- Cmd{1, TurnData{coords, result}}
				break
			}
			board.draw()

		case kCmdWaitForOpponent:
			println("Waiting for opponent...")

		case kCmdWaitForResultConfirmation:
			println("Waiting for game result confirmation with the peer...")

		case kCmdGameFinished:
			var result = board.gameResult

			board.draw()
			println()
			switch result {
			case Draw:
				fmt.Println("*** \x1b[7mIt's a draw\x1b[0m ***")
			case MeWin:
				fmt.Println("*** \x1b[42m\x1b[30mYou win!\x1b[0m ***")
			case HeWin:
				fmt.Println("*** \x1b[41m\x1b[30mYou lose!\x1b[0m ***")
			}
			os.Exit(0)
		}
	}

	/////////////////////////////////////////////////////////////////////

	/*for {*/
		/*if !myTurn {*/
			/*println("Waiting for opponent...")*/
			/*result, err := board.waitForOpponent()*/
			/*if err != nil {*/
				/*printError(err)*/
				/*os.Exit(0)*/
			/*}*/
			/*checkResult(result)*/
			/*continue*/
		/*}*/

		/*println("\n<<< \x1b[1mYour turn\x1b[0m >>>")*/

		/*for {*/
			/*board.draw()*/

			/*var move, err = getMove()*/
			/*if err != nil {*/
				/*os.Exit(0)*/
			/*}*/

			/*coords, err := parseMove(move)*/
			/*if err != nil {*/
				/*printError(err)*/
				/*continue*/
			/*}*/

			/*result, err := board.makeMove(coords, ownChar)*/
			/*if err != nil {*/
				/*printError(err)*/
				/*continue*/
			/*}*/
			/*checkResult(result)*/

			/*break*/
		/*}*/
	/*}*/
}
