package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

var stdin = bufio.NewReader(os.Stdin)

func getMove() (str string, err error) {
	print("> ")

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

	if !(parseLetter(move[0], &coords[0]) || parseLetter(move[1], &coords[0])) {
		return
	}
	if !(parseDigit(move[0], &coords[1]) || parseDigit(move[1], &coords[1])) {
		return
	}

	return coords, nil
}

func printError(err error) {
	fmt.Printf("\x1b[31m%v\x1b[0m\n", err)
}

func main() {
	fmt.Println("*** Welcome to Tic-Tac-Go ***")
	var board = NewBoard()
	ownChar, oppChar = 'X', 'O'

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
			os.Exit(0)
		}
	}

	for {
		println("\nYour turn.")

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

			result, err := board.makeMove(coords, 'X')
			if err != nil {
				printError(err)
				continue
			}
			checkResult(result)

			println("Waiting for opponent...")
			result, err = board.waitForOpponent()
			if err != nil {
				printError(err)
				os.Exit(0)
			}
			checkResult(result)

			break
		}
	}
}
