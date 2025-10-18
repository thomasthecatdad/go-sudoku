package solver

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"golang.org/x/sync/errgroup"
)

var (
	errInvalidValue   = errors.New("attempted to set cell to invalid value")
	errPopulatedValue = errors.New("failed to set cell to value, cell was already populated")
	errBannedValue    = errors.New("attempted to set cell to banned value")
	errBuildFailure   = errors.New("failed to build the board")
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
		Speaker:      make(chan int, 24),
		bannedValues: map[int]bool{},
	}

	return &c
}

func (c *Cell) SetValue(value int) error {
	if value < 1 || value > 9 {
		return fmt.Errorf("%w [%d]", errInvalidValue, value)
	}

	if c.value != 0 {
		return fmt.Errorf("%w [%d, %d]", errPopulatedValue, value, c.value)
	}

	if c.bannedValues[value] {
		return fmt.Errorf("%w [%d]", errBannedValue, value)
	}

	c.value = value

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

func (c *Cell) GetValue() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

func (c *Cell) Listen() error {
	if c.value != 0 {
		return nil
	}
	for rec := range c.Speaker {
		c.mu.Lock()
		c.bannedValues[rec] = true
		if len(c.bannedValues) == 8 {
			for i := 1; i <= 9; i++ {
				if !c.bannedValues[i] {
					err := c.SetValue(i)
					c.mu.Unlock()
					return err
				}
			}
		}
		c.mu.Unlock()
	}
	return fmt.Errorf("should not get here")
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

func PopulateBoard(board *[9][9]*Cell, values [][]string) error {
	// Initialize Board
	for rowNum := range board {
		for colNum := range board[rowNum] {
			fmt.Printf("x: %d, y: %d\n", rowNum, colNum)
			num, err := strconv.Atoi(values[rowNum][colNum])
			if err != nil {
				return fmt.Errorf("%w: %w", errBuildFailure, err)
			}
			if num == 0 {
				continue
			}
			err = board[rowNum][colNum].SetValue(num)
			if err != nil {
				return fmt.Errorf("%w: %w", errBuildFailure, err)
			}
		}
	}

	fmt.Println("finished initializing board")

	return nil
}

func VisualizeBoard(board *[9][9]*Cell) {
	rowSep := "-----------------------------------------"
	fmt.Println(rowSep)
	for rowNum := range board {
		if rowNum%3 == 0 {
			fmt.Println(rowSep)
		}
		fmt.Print("|")
		for colNum := range board[rowNum] {
			if colNum%3 == 0 {
				fmt.Printf("|")
			}
			val := board[rowNum][colNum].GetValue()
			if val != 0 {
				fmt.Printf(" %d |", val)
			} else {
				fmt.Printf("   |")
			}

		}
		fmt.Println("|")
		fmt.Println(rowSep)

	}
	fmt.Println(rowSep)
}

func SolveBoard(board *[9][9]*Cell) {
	g := new(errgroup.Group)

	for rowNum := range board {
		for colNum := range board[rowNum] {
			g.Go(board[rowNum][colNum].Listen)
		}
	}

	err := g.Wait()
	if err != nil {
		fmt.Printf("error occured while solving board: %v", err)
	}
}
