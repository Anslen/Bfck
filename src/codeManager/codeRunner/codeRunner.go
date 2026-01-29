package coderunner

import (
	"fmt"
	"slices"

	"github.com/Anslen/Bfck/codeManager/code"
	"github.com/Anslen/Bfck/memory"
)

type ReturnCode byte

const (
	ReturnFinish = iota
	ReturnBreakPoint
	ReturnStep // Step running finish
)

type CodeRunner struct {
	code             *code.Code
	codeIndex        int // Point at next operator to execute
	memory           *memory.Memory
	debugFlag        bool
	breakPoint       []uint64
	codeBreakPointed []bool
}

func New(code *code.Code, debugFlag bool) (ret *CodeRunner) {
	if code == nil {
		panic("CodeRunner: code is nil")
	}

	if debugFlag {
		ret = &CodeRunner{
			code:             code,
			codeIndex:        0,
			memory:           memory.New(),
			debugFlag:        true,
			breakPoint:       make([]uint64, 0),
			codeBreakPointed: make([]bool, code.CodeCount),
		}
	} else {
		ret = &CodeRunner{
			code:             code,
			codeIndex:        0,
			memory:           memory.New(),
			debugFlag:        false,
			breakPoint:       nil,
			codeBreakPointed: nil,
		}
	}
	return
}

// AddBreakPoint adds a breakpoint at the specified line.
func (cr *CodeRunner) AddBreakPoint(line uint64) (err error) {
	if !cr.debugFlag {
		panic("CodeRunner: can't add breakpoint when not in debug mode")
	}

	// Check line range
	if line == 0 || line > cr.code.LineCount {
		err = fmt.Errorf("Error: breakpoint out of range, line count is %v, get line %v", cr.code.LineCount, line)
		return
	}

	// Find position and add breakpoint
	position, found := slices.BinarySearch(cr.breakPoint, line)
	if !found {
		// Warn if breakpoint at empty line
		if cr.code.LineBegins[line-1] == -1 {
			fmt.Printf("Warning: breakpoint at line %v will not work\n\n", line)
			return
		}

		cr.codeBreakPointed[cr.code.LineBegins[line-1]] = true
		cr.breakPoint = slices.Insert(cr.breakPoint, position, line)
		// Pirnt add success information
		fmt.Printf("Breakpoint added at line %v\n\n", line)

	} else {
		err = fmt.Errorf("Warning: Breakpoint at line %v already existed\n\n", line)
	}

	return
}

// RemoveBreakPoint removes the breakpoint at the specified index.
//
// CAUSION: index start from 1
func (cr *CodeRunner) RemoveBreakPoint(index int) (err error) {
	if !cr.debugFlag {
		panic("CodeRunner: can't remove breakpoint when not in debug mode")
	}

	// Check index range
	if index <= 0 || index > len(cr.breakPoint) {
		err = fmt.Errorf("Error: breakpoint index out of range, get %v, breakpoint count is %v", index, len(cr.breakPoint))
		return
	}

	var removedLine uint64 = cr.breakPoint[index-1]
	var removedCodeIndex int = cr.code.LineBegins[removedLine-1]
	cr.breakPoint = slices.Delete(cr.breakPoint, index-1, index)

	// Print remove success information
	fmt.Printf("Breakpoint %v removed\n\n", index)

	// Remove code breakpoint mark if nessesary
	// Check last breakpoint
	if index != 1 && cr.code.LineBegins[cr.breakPoint[index-2]-1] == removedCodeIndex {
		return
	}
	// Check next breakpoint
	if index != len(cr.breakPoint)+1 && cr.code.LineBegins[cr.breakPoint[index-1]-1] == removedCodeIndex {
		return
	}

	cr.codeBreakPointed[removedCodeIndex] = false
	return
}

// ClearBreakPoints removes all breakpoints.
func (cr *CodeRunner) ClearBreakPoints() {
	if !cr.debugFlag {
		panic("CodeRunner: can't clear breakpoints when not in debug mode")
	}
	cr.breakPoint = make([]uint64, 0)
	for index := range cr.codeBreakPointed {
		cr.codeBreakPointed[index] = false
	}
	fmt.Print("All breakpoints cleared\n\n")
}

// PrintBreakPoint prints all breakpoints.
func (cr *CodeRunner) PrintBreakPoint() {
	if !cr.debugFlag {
		panic("CodeRunner: can't print breakpoints when not in debug mode")
	}

	// Check if any breakpoint exists
	if len(cr.breakPoint) == 0 {
		fmt.Print("No breakpoints exist now.\n\n")
		return
	}

	// Print each breakpoint
	fmt.Println("Num\tLine")
	for index, line := range cr.breakPoint {
		fmt.Printf("%v\t%v\n", index+1, line)
	}
	fmt.Print("\n")
}

// PrintCode prints analysed code information.
func (cr *CodeRunner) PrintAllOperator() {
	cr.code.PrintAll()
}

// PrintNextOperator prints the next operator to be executed.
func (cr *CodeRunner) PrintNextOperator() {
	if cr.codeIndex < 0 || cr.codeIndex >= cr.code.CodeCount {
		panic("CodeRunner: code index out of range")
	}

	cr.code.Print(cr.codeIndex)
}

// PeekBytes peeks bytes from memory with the given offset and length.
//
// Offset is relative to the current memory pointer.
func (cr *CodeRunner) PeekBytes(offset, length int) (ret []byte) {
	return cr.memory.PeekBytes(offset, length)
}

// Run starts running the code from the beginning.
//
// Return ReturnCode and current line number (1-based), line only valid when hit breakpoint
func (cr *CodeRunner) Run() (ret ReturnCode) {
	cr.codeIndex = 0
	cr.memory = memory.New()
	ret = cr.Continue()
	return
}

// Continue continues running the code from the current position.
//
// Return ReturnCode and current line number (1-based), line only valid when hit breakpoint
func (cr *CodeRunner) Continue() (ret ReturnCode) {
	for cr.codeIndex < cr.code.CodeCount {
		// Check for breakpoint
		if cr.debugFlag && cr.codeBreakPointed[cr.codeIndex] {
			// Hit breakpoint
			return ReturnBreakPoint
		}

		// Execute operator
		cr.executeOperator()
	}

	cr.codeIndex = 0
	return ReturnFinish
}

// Step executes the next operator, ignore breakpoints.
//
// Return ReturnCode and current line number (1-based), line will be not setted
func (cr *CodeRunner) Step() (ret ReturnCode) {
	// Check if code has finished
	if cr.codeIndex >= cr.code.CodeCount {
		panic("CodeRunner: code index out of range")
	}

	cr.executeOperator()

	// Check code finish
	if cr.codeIndex >= cr.code.CodeCount {
		cr.codeIndex = 0
		return ReturnFinish
	} else {
		return ReturnStep
	}
}

// executeOperator executes the current operator and advances the code index.
func (cr *CodeRunner) executeOperator() {
	var operator code.Operator = cr.code.Operators[cr.codeIndex]
	var auxiliary uint64 = cr.code.Auxiliary[cr.codeIndex]
	cr.codeIndex++

	switch operator {
	case code.OpAdd:
		cr.memory.Add(auxiliary)

	case code.OpSub:
		cr.memory.Sub(auxiliary)

	case code.OpMoveLeft:
		// Memory block may change after moving pointer
		cr.memory = cr.memory.MovePtr(-int(auxiliary))

	case code.OpMoveRight:
		// Memory block may change after moving pointer
		cr.memory = cr.memory.MovePtr(int(auxiliary))

	case code.OpLeftBracket:
		if cr.memory.Peek(0) == 0 {
			cr.codeIndex = int(auxiliary)
		}

	case code.OpRightBracket:
		if cr.memory.Peek(0) != 0 {
			cr.codeIndex = int(auxiliary)
		}

	case code.OpInput:
		var input rune
		fmt.Scanf("%c", &input)
		cr.memory.Poke(byte(input))

	case code.OpOutput:
		fmt.Printf("%c", cr.memory.Peek(0))
	}
}
