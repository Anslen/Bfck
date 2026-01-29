package code

import "fmt"

type Operator byte

const (
	OpAdd = iota
	OpSub
	OpMoveLeft
	OpMoveRight
	OpLeftBracket
	OpRightBracket
	OpInput
	OpOutput
	Invalid // Only for internal use
)

// code represents the analysed Brainfuck code.
//
// lineBegins will be nil if not in debug mode.
//
// If line is empty, lineBegins will store next valid position.
//
// If no next valid position, lineBegins will store -1.
type Code struct {
	Operators  []Operator
	Auxiliary  []uint64 // Auxiliary data for operators, times for +/- and moves, jump positions for brackets, 1 for i/o
	CodeCount  int
	LineCount  uint64 // Number of lines in the original code
	LineBegins []int  // Begin index for each line
}

func New(debugFlag bool) (ret *Code) {
	ret = &Code{
		Operators:  make([]Operator, 0),
		Auxiliary:  make([]uint64, 0),
		CodeCount:  0,
		LineCount:  0,
		LineBegins: nil,
	}
	if debugFlag {
		ret.LineBegins = make([]int, 0)
	}
	return
}

// Print prints the code with auxiliary data and line begins (if any).
func (c *Code) Print() {
	fmt.Printf("\nTotal operators count: %v\n\n", c.CodeCount)

	fmt.Println("Code with auxiliary:")
	for index := 0; index < c.CodeCount; index++ {
		fmt.Printf("%-4d %-10s %d\n", index, c.Operators[index].String(), c.Auxiliary[index])
	}

	fmt.Printf("\nLines count: %v\n\n", c.LineCount)
	if c.LineBegins != nil {
		fmt.Println("Line begins at:")
		fmt.Println("Line\tbegin")
		for index, line := range c.LineBegins {
			fmt.Printf("%d\t%d\n", index+1, line)
		}
	}
}

func (op Operator) String() string {
	switch op {
	case OpAdd:
		return "Add"

	case OpSub:
		return "Sub"

	case OpMoveLeft:
		return "MoveLeft"

	case OpMoveRight:
		return "MoveRight"

	case OpLeftBracket:
		return "LeftBracket"

	case OpRightBracket:
		return "RightBracket"

	case OpInput:
		return "Input"

	case OpOutput:
		return "Output"
	}
	return "Invalid"
}
