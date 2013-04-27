package main

import (
	"errors"
	"fmt"
)

var ownChar, oppChar int

const (
	NoMove = iota
	OKMove

	GameFinished
	MeWin
	HeWin
	Draw
)

func formatCell(char int) string {
	var seq string
	if char == oppChar {
		seq = "\x1b[41m[%c]\x1b[0m"
	} else if char == ownChar {
		seq = "\x1b[7m[%c]\x1b[0m"
	} else {
		seq = "[%c]"
	}
	return fmt.Sprintf(seq, char)
}

func mapChar(char int) int {
	if char == ownChar {
		return MeWin
	}
	return HeWin
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
	fmt.Printf("\n   1   2   3\na %s %s %s\nb %s %s %s\nc %s %s %s\n",
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
		return mapChar(char)
	}

	// Check vertical
	char = b.b[0][j]
	if char != ' ' && b.b[1][j] == char && b.b[2][j] == char {
		return mapChar(char)
	}

	// Check diagonals
	if (i+j)%2 == 1 {
		return 0 // not a diagonal
	}

	char = b.b[0][0]
	if char != ' ' && b.b[1][1] == char && b.b[2][2] == char {
		return mapChar(char)
	}

	char = b.b[0][2]
	if char != ' ' && b.b[1][1] == char && b.b[2][0] == char {
		return mapChar(char)
	}

	return 0
}

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
				return b.makeMove([2]int{i, j}, oppChar)
			}
		}
	}
	return NoMove, errors.New("No free cell found")
}
