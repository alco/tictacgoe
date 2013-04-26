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
	freeCells int
}

func NewBoard() *Board {
	var b = Board{}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			b.b[i][j] = ' '
		}
	}
	b.freeCells = 9
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

const (
	OKMove = iota
	NoMove
	GameFinished

	MeWin
	HeWin
	Draw
)

func (b *Board) makeMove(coords [2]int, char int) (int, error) {
	var cell = &b.b[coords[0]][coords[1]]
	if *cell == ' ' {
		*cell = 'X'
		b.freeCells -= 1
		if b.freeCells == 0 {
			return Draw, nil
		} else {
			return OKMove, nil
		}
	}
	return NoMove, errors.New("\x1b[31mCell already taken.\x1b[0m")
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

func parseLetter(char byte, coord *int) bool {
	switch char {
	case 'a':
		*coord = 0
	case 'b':
		*coord = 1
	case 'c':
		*coord = 2
	default:
		return false
	}
	return true
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

func main() {
	fmt.Println("*** Welcome to Tic-Tac-Go ***")
	var board = NewBoard()

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
				fmt.Printf("\x1b[31m%v\x1b[0m\n", err)
				continue
			}
			result, err := board.makeMove(coords, 'X')
			if err != nil {
				fmt.Println(err)
				continue
			} else if result > GameFinished {
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

			break
		}
	}
}
