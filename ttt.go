package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Board struct {
	b [3][3]int
}

func NewBoard() *Board {
	var b = Board{}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			b.b[i][j] = ' '
		}
	}
	return &b
}

func (b *Board) draw() {
	fmt.Printf(`
   1   2   3
a [%c] [%c] [%c]
b [%c] [%c] [%c]
c [%c] [%c] [%c]
`,
		b.b[0][0], b.b[0][1], b.b[0][2],
		b.b[1][0], b.b[1][1], b.b[1][2],
		b.b[2][0], b.b[2][1], b.b[2][2])

}

func printPrompt() {
	print("> ")
}

var stdin = bufio.NewReader(os.Stdin)

func getMove() (str string, err error) {
	printPrompt()

	str, err = stdin.ReadString('\n')
	if err != nil {
		return
	}
	return str[:len(str)-1], err
}

func parseLetter(str string, coord *int) bool {
	switch str {
	case "a":
		*coord = 0
	case "b":
		*coord = 1
	case "c":
		*coord = 2
	default:
		return false
	}
	return true
}

func parseDigit(str string, coord *int) bool {
	if len(str) == 1 && str[0] >= '1' && str[0] <= '3' {
		*coord = int(str[0] - '1')
		return true
	}
	return false
}

func parseMove(move string) (coords [2]int, err error) {
	err = errors.New("err: Invalid move.")

	if len(move) == 0 {
		err = errors.New("err: Please make a move.")
		return
	}

	var fields = strings.Fields(move)
	if len(fields) != 2 {
		return
	}

	if !(parseLetter(fields[0], &coords[0]) || parseLetter(fields[1], &coords[0])) {
		return
	}
	if !(parseDigit(fields[0], &coords[1]) || parseDigit(fields[1], &coords[1])) {
		return
	}

	return coords, nil
}

func main() {
	fmt.Println("*** Welcome to Tic-Tac-Go ***")
	var board = NewBoard()

	for {
		board.draw()
		println("\nYour turn.")

		for {
			var move, err = getMove()
			if err != nil {
				os.Exit(0)
			}

			coords, err := parseMove(move)
			if err != nil {
				fmt.Printf("\x1b[31m%v\x1b[0m\n", err)
				board.draw()
				continue
			}

			var cell = &board.b[coords[0]][coords[1]]
			if *cell == ' ' {
				*cell = 'X'
			} else {
				fmt.Println("\x1b[31mCell already taken.\x1b[0m")
				board.draw()
				continue
			}

			break
		}
	}
}
