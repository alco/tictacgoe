package main

import (
	"bufio"
	"fmt"
	"os"
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
			board.b[0][0] = 'X'
			println(move)
			break
		}
	}
}
