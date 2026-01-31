/*
 * Copyright (C) 2026 Anslen
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

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

// PrintAll prints all code with auxiliary data and line begins (if any).
func (c *Code) PrintAll() {
	fmt.Printf("\nTotal operators count: %v\n\n", c.CodeCount)

	fmt.Println("Code with auxiliary:")
	var loopCount uint64 = 0
	var loopCountStack []uint64 = make([]uint64, 0)

	for index, operator := range c.Operators {
		// Print loop labels
		if operator == OpLeftBracket {
			loopCount++
			loopCountStack = append(loopCountStack, loopCount)
			fmt.Printf("L%v:\n", loopCount)
		}
		fmt.Printf("  %-8d %-15s %d\n", index, operator.String(), c.Auxiliary[index])
		// Print loop end labels
		if operator == OpRightBracket {
			if len(loopCountStack) == 0 {
				panic("Code: unmatched right bracket when printing")
			}
			var lastLoopCount uint64 = loopCountStack[len(loopCountStack)-1]
			fmt.Printf("L%v End\n", lastLoopCount)
			loopCountStack = loopCountStack[:len(loopCountStack)-1]
		}
	}

	fmt.Printf("\nLines count: %v\n\n", c.LineCount)
	if c.LineBegins != nil {
		fmt.Println("Line begins at:")
		fmt.Println("Line\tbegin")
		for index, line := range c.LineBegins {
			fmt.Printf("  %d\t%d\n", index+1, line)
		}
	}
}

// Print prints the operator and auxiliary data at the given index.
func (c Code) Print(index int) {
	if index < 0 || index >= c.CodeCount {
		panic("Code: code index out of range")
	}

	fmt.Printf("%-8d %-15s %d\n", index, c.Operators[index].String(), c.Auxiliary[index])
}

// ToOperator converts a rune character to the corresponding Operator.
func ToOperator(char rune) (ret Operator) {
	switch char {
	case '+':
		return OpAdd

	case '-':
		return OpSub

	case '<':
		return OpMoveLeft

	case '>':
		return OpMoveRight

	case '[':
		return OpLeftBracket

	case ']':
		return OpRightBracket

	case '.':
		return OpOutput

	case ',':
		return OpInput
	}
	return Invalid
}

// Reverse returns the opposite operator of the given operator.
func (op Operator) Reverse() (ret Operator) {
	switch op {
	case OpAdd:
		return OpSub

	case OpSub:
		return OpAdd

	case OpMoveLeft:
		return OpMoveRight

	case OpMoveRight:
		return OpMoveLeft
	}
	return Invalid
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
