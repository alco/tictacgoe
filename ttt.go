package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net/rpc"
	"os"
	"strings"
	"time"
)

var stdin = bufio.NewReader(os.Stdin)

func getMove() (str string, err error) {
	print("> ") // prompt

	str, err = stdin.ReadString('\n')
	if err != nil {
		return
	}
	return str[:len(str)-1], err
}

func parseLetter(char byte, coord *int) bool {
	if char >= 'a' && char <= 'c' {
		*coord = int(char - 'a')
		return true
	}
	return false
}

func parseDigit(char byte, coord *int) bool {
	if char >= '1' && char <= '3' {
		*coord = int(char - '1')
		return true
	}
	return false
}

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

func parseCommand(str string) (string, error) {
	return "", errors.New("err: invalid command")
}

func printError(err interface{}) {
	fmt.Printf("\x1b[31m%v\x1b[0m\n", err)
}

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
	var board = NewBoard()

	rand.Seed(time.Now().Unix())

	var firstPlayer int
	var done chan bool
	var client *rpc.Client
	if serverMode {
		fmt.Printf("Listening on port %v...\n", *port)
		firstPlayer, done = listen(board, *port)
	} else {
		println("Connecting to server...")
		firstPlayer, client = connectToServer(fmt.Sprintf("%v:%v", *addr, *port))
	}

	var checkResult = func(result int) {
		if result > GameFinished {
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
			err := rpc_finishGameWithResult(client, result)
			if err != nil {
				printError(err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	var myTurn bool
	if firstPlayer == 1 {
		ownChar, oppChar = 'X', 'O'
		myTurn = true
	} else {
		ownChar, oppChar = 'O', 'X'
		myTurn = false
	}

	if serverMode {
		// When using RPC, the client drives the game. This feels weird.
		// So we are just waiting for the client to call RPC methods in another
		// goroutine. Once the game is finished, we receive a message on the
		// 'done' channel.
		if !myTurn {
			println()
			println("Waiting for opponent...")
			board.draw()
		}

		<-done
		os.Exit(0)
	} else {
		if !myTurn {
			// Let the server make the initial move
			board.draw()
			println()
			println("Waiting for opponent...")

			coords, serverResult, err := rpc_waitForOpponent(client)
			if err != nil {
				printError(err)
				os.Exit(1)
			}
			result, err := board.makeMove(coords, oppChar)
			if err != nil {
				printError(err)
				os.Exit(1)
			}
			if invert(serverResult) != result {
				printError(errors.New("Either party is cheating!"))
				os.Exit(1)
			}
		}

		for {
			println("\n<<< \x1b[1mYour turn\x1b[0m >>>")

			var result int
			var coords [2]int
			for {
				board.draw()

				var move, err = getMove()
				if err != nil {
					os.Exit(0)
				}

				coords, err = parseMove(move)
				if err != nil {
					printError(err)
					continue
				}

				result, err = board.makeMove(coords, ownChar)
				if err != nil {
					printError(err)
					continue
				}

				break
			}
			serverResult, err := rpc_makeMove(client, coords)
			if err != nil {
				printError(err)
				os.Exit(1)
			}
			if invert(serverResult) != result {
				printError(errors.New("Either party is cheating!"))
				os.Exit(1)
			}
			checkResult(result)

			board.draw()
			println()
			println("Waiting for opponent...")

			coords, serverResult, err = rpc_waitForOpponent(client)
			if err != nil {
				printError(err)
				os.Exit(1)
			}
			result, err = board.makeMove(coords, oppChar)
			if err != nil {
				printError(err)
				os.Exit(1)
			}
			if invert(serverResult) != result {
				printError(errors.New("Either party is cheating!"))
				os.Exit(1)
			}
			checkResult(result)
		}

		os.Exit(0)
	}

	for {
		if !myTurn {
			println("Waiting for opponent...")
			result, err := board.waitForOpponent()
			if err != nil {
				printError(err)
				os.Exit(0)
			}
			checkResult(result)
			continue
		}

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
			checkResult(result)

			break
		}
	}
}
