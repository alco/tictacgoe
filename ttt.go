package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

func formatCell(char int) string {
	var seq string
	if char == 'O' {
		seq = "\x1b[41m[%c]\x1b[0m"
	} else if char == 'X' {
		seq = "\x1b[7m[%c]\x1b[0m"
	} else {
		seq = "[%c]"
	}
	return fmt.Sprintf(seq, char)
}

type Board struct {
	b         [3][3]int
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
a %s %s %s
b %s %s %s
c %s %s %s
`,
		formatCell(b.b[0][0]), formatCell(b.b[0][1]), formatCell(b.b[0][2]),
		formatCell(b.b[1][0]), formatCell(b.b[1][1]), formatCell(b.b[1][2]),
		formatCell(b.b[2][0]), formatCell(b.b[2][1]), formatCell(b.b[2][2]))

}

func (b *Board) checkWinningCondition(coords [2]int) int {
	var i, j = coords[0], coords[1]
	var char int

	// Check horizontal
	char = b.b[i][0]
	if char != ' ' && b.b[i][1] == char && b.b[i][2] == char {
		return MeWin
	}

	// Check vertical
	char = b.b[0][j]
	if char != ' ' && b.b[1][j] == char && b.b[2][j] == char {
		return MeWin
	}

	// Check diagonals
	if (i+j)%2 == 1 {
		return 0 // not a diagonal
	}

	char = b.b[0][0]
	if char != ' ' && b.b[1][1] == char && b.b[2][2] == char {
		return MeWin
	}

	char = b.b[0][2]
	if char != ' ' && b.b[1][1] == char && b.b[2][0] == char {
		return MeWin
	}

	return 0
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
		*cell = char
		b.freeCells -= 1
		if result := b.checkWinningCondition(coords); result != 0 {
			return result, nil
		} else if b.freeCells == 0 {
			return Draw, nil
		}
		return OKMove, nil
	}
	return NoMove, errors.New("Cell already taken.")
}

func (b *Board) waitForOpponent() (int, error) {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if b.b[i][j] == ' ' {
				return b.makeMove([2]int{i, j}, 'O')
			}
		}
	}
	return NoMove, errors.New("No free cell found")
}

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

func printError(err error) {
	fmt.Printf("\x1b[31m%v\x1b[0m\n", err)
}

func main() {
	fmt.Println("*** Welcome to Tic-Tac-Go ***")
	var board = NewBoard()

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
