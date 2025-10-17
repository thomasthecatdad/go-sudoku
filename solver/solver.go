package solver

import (
	"errors"
	"fmt"
	"sync"
)

var (
	errInvalidValue   = errors.New("attempted to set cell to invalid value")
	errPopulatedValue = errors.New("failed to set cell to value, cell was already populated")
	errBannedValue    = errors.New("attempted to set cell to banned value")
)

type Cell struct {
	value        int
	X            int
	Y            int
	Block        []chan<- int
	Row          []chan<- int
	Column       []chan<- int
	Speaker      chan int
	bannedValues map[int]bool
	mu           sync.Mutex
}

func NewCell() *Cell {
	c := Cell{
		Speaker: make(chan int, 24),
	}

	return &c
}

func (c *Cell) SetValue(value int) error {
	if value < 1 || value > 9 {
		return fmt.Errorf("%w [%d]", errInvalidValue, value)
	}

	c.mu.Lock()

	if c.value != 0 {
		return fmt.Errorf("%w [%d, %d]", errPopulatedValue, value, c.value)
	}

	if c.bannedValues[value] {
		return fmt.Errorf("%w [%d]", errBannedValue, value)
	}

	c.value = value
	c.mu.Unlock()

	for _, cellSpeaker := range c.Block {
		cellSpeaker <- value
	}

	for _, cellSpeaker := range c.Row {
		cellSpeaker <- value
	}

	for _, cellSpeaker := range c.Column {
		cellSpeaker <- value
	}

	return nil
}

type coords struct {
	X int
	Y int
}

/* CreateBoard returns a 9x9 2D array of Cells
 * Starting from the top left of the Sudoku board, cells are indexed like so
 *  ----------------------------
 *  | [0][0] | [0][1] | [0][2] |
 *  ----------------------------
 *  | [1][0] | [1][1] | [1][2] |
 *  ----------------------------
 *  | [2][0] | [2][1] | [2][2] |
 *  ----------------------------
 */
func CreateBlankBoard() *[9][9]*Cell {
	board := [9][9]*Cell{}
	column := [9][9]*Cell{}
	block := [9][]*Cell{}

	// Initialize Board
	for rowNum := range board {
		for colNum := range board[rowNum] {
			cell := NewCell()
			board[rowNum][colNum] = cell
			column[colNum][rowNum] = cell
			blockNum := (rowNum/3)*3 + (colNum / 3)
			block[blockNum] = append(block[blockNum], cell)
		}
	}

	// Connect Cells
	for _, r := range board {
		connectRows(r)
	}
	for _, c := range column {
		connectCols(c)
	}
	for _, b := range block {
		connectBlocks(b)
	}

	return &board
}

func connectRows(cells [9]*Cell) {
	for speakerNum := range cells {
		for listenerNum := range cells {
			if speakerNum == listenerNum {
				continue
			}
			cells[listenerNum].Row = append(cells[listenerNum].Row, cells[speakerNum].Speaker)
		}
	}
}

func connectCols(cells [9]*Cell) {
	for speakerNum := range cells {
		for listenerNum := range cells {
			if speakerNum == listenerNum {
				continue
			}
			cells[listenerNum].Column = append(cells[listenerNum].Column, cells[speakerNum].Speaker)
		}
	}
}

func connectBlocks(cells []*Cell) {
	for speakerNum := range cells {
		for listenerNum := range cells {
			if speakerNum == listenerNum {
				continue
			}
			cells[listenerNum].Block = append(cells[listenerNum].Block, cells[speakerNum].Speaker)
		}
	}
}
