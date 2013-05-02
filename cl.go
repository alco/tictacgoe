package main

// Command line interface for the TicTagGoe program.

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

// Format the current board state nicely
func drawBoard(b *Board) {
	var formatCell = func(char int) string {
		var seq string
		if char == b.oppChar {
			seq = "\x1b[41m[%c]\x1b[0m"
		} else if char == b.ownChar {
			seq = "\x1b[7m[%c]\x1b[0m"
		} else {
			seq = "[%c]"
		}
		return fmt.Sprintf(seq, char)
	}
	fmt.Printf("\n   1   2   3\na %s %s %s\nb %s %s %s\nc %s %s %s\n",
		formatCell(b.b[0][0]), formatCell(b.b[0][1]), formatCell(b.b[0][2]),
		formatCell(b.b[1][0]), formatCell(b.b[1][1]), formatCell(b.b[1][2]),
		formatCell(b.b[2][0]), formatCell(b.b[2][1]), formatCell(b.b[2][2]))

}

// CLI flags
var mode = flag.String("mode", "server", "Which mode to run in: server or client")
var addr = flag.String("addr", "localhost", "Address to connect to")
var port = flag.Int("port", 8888, "Port to listen on or connect to")

func runCommandLine() {
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

	var net = NewNet()

	if serverMode {
		fmt.Printf("Listening on port %v...\n", *port)
		err := net.Listen(*port)
		if err != nil {
			printError(err)
			os.Exit(1)
		}
	} else {
		println("Connecting to server...")
		err := net.ConnectToServer(fmt.Sprintf("%v:%v", *addr, *port))
		if err != nil {
			printError(err)
			os.Exit(1)
		}
	}

	println("\n*** \x1b[7mGame started\x1b[0m ***")

	for {
		switch net.GetCommand() {
		case kCmdMakeTurn:
			println("\n<<< \x1b[1mYour turn\x1b[0m >>>")

			for {
				drawBoard(net.Board)

				var move, err = getMove()
				if err != nil {
					os.Exit(0)
				}

				coords, err := parseMove(move)
				if err != nil {
					printError(err)
					continue
				}

				result, err := net.Board.makeMove(coords, net.Board.ownChar)
				if err != nil {
					printError(err)
					continue
				}

				net.SendResponse(Cmd{1, TurnData{coords, result}})
				break
			}

		case kCmdWaitForOpponent:
			drawBoard(net.Board)
			println("Waiting for opponent...")

		case kCmdWaitForResultConfirmation:
			drawBoard(net.Board)
			println("Waiting for game result confirmation with the peer...")

		case kCmdGameFinished:
			drawBoard(net.Board)
			println()

			switch net.GameResult {
			case kGameResultDraw:
				println("*** \x1b[7mIt's a draw\x1b[0m ***")
			case kGameResultMeWin:
				println("*** \x1b[42m\x1b[30mYou win!\x1b[0m ***")
			case kGameResultHeWin:
				println("*** \x1b[41m\x1b[30mYou lose!\x1b[0m ***")
			}
			os.Exit(0)
		}
	}
}
