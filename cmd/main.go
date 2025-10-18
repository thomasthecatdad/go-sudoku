package main

import (
	_ "embed"
	"encoding/csv"
	"os"

	"github.com/thomasthecatdad/go-sudoku/solver"
)

func main() {
	f, err := os.Open("boards/easy.csv")
	if err != nil {
		panic(err)
	}
	r := csv.NewReader(f)

	board := solver.CreateBlankBoard()

	boardStrings, err := r.ReadAll()
	if err != nil {
		panic(err)
	}

	err = solver.PopulateBoard(board, boardStrings)
	if err != nil {
		panic(err)
	}
	solver.VisualizeBoard(board)
	solver.SolveBoard(board)
	solver.VisualizeBoard(board)
}
