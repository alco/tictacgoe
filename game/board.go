package game

// Game logic implementation

import (
	"errors"
)

const (
	Player1Char = 'X'
	Player2Char = 'O'
)

const (
	NoMove = iota
	OKMove

	GameFinished
	Player1Win
	Player2Win
	Draw
)

func mapChar(char int) int {
	if char == Player1Char {
		return Player1Win
	}
	return Player2Win
}

type Board struct {
	b           [3][3]int
	freeCells   int
	finalResult int
	ownChar     int
	oppChar     int
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

/// Read-only getters

func (b *Board) OwnChar() int {
	return b.ownChar
}

func (b *Board) OppChar() int {
	return b.oppChar
}

func (b *Board) At(i, j int) int {
	return b.b[i][j]
}

func (b *Board) FinalResult() int {
	return b.finalResult
}

/// Public methods

func (b *Board) SetFirstPlayer(firstPlayer int) {
	if firstPlayer == 0 {
		b.ownChar, b.oppChar = Player1Char, Player2Char
	} else {
		b.ownChar, b.oppChar = Player2Char, Player1Char
	}
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

func (b *Board) MakeOwnMove(coords [2]int) (int, error) {
	return b.makeMove(coords, b.ownChar)
}

func (b *Board) MakeOppMove(coords [2]int) (int, error) {
	return b.makeMove(coords, b.oppChar)
}

func (b *Board) makeMove(coords [2]int, char int) (int, error) {
	var cell = &b.b[coords[0]][coords[1]]
	if *cell == ' ' {
		*cell = char
		b.freeCells -= 1
		if result := b.checkWinningCondition(coords); result != 0 {
			b.finalResult = result
			return result, nil
		} else if b.freeCells == 0 {
			b.finalResult = Draw
			return Draw, nil
		}
		return OKMove, nil
	}
	return NoMove, errors.New("Cell already taken.")
}

// Used for testing
func (b *Board) MakeAIMove() (int, error) {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if b.b[i][j] == ' ' {
				return b.makeMove([2]int{i, j}, b.oppChar)
			}
		}
	}
	return NoMove, errors.New("No free cell found")
}
