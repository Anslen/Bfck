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

package coderunner

import (
	"fmt"
	"slices"

	"github.com/Anslen/Bfck/codeManager/code"
	"github.com/Anslen/Bfck/memory"
)

type ReturnCode byte

const (
	ReturnAfterFinish = iota
	ReturnAfterStep   // Step running finish
	ReturnReachBreakPoint
	ReturnReachWatch
	ReturnReachUntil
	returnAfterExecuteOperator // For internal function executeOperator
)

type CodeRunner struct {
	code             *code.Code
	codeIndex        int // Point at next operator to execute
	memory           *memory.Memory
	debugFlag        bool
	breakPoint       []uint64
	codeBreakPointed []bool
	breakPointUsed   bool
	isWatching       bool
	watchOffset      int
	untilStatus      bool
}

func New(code *code.Code, debugFlag bool) (ret *CodeRunner) {
	if code == nil {
		panic("CodeRunner: code is nil")
	}

	if debugFlag {
		ret = &CodeRunner{
			code:             code,
			memory:           memory.New(),
			debugFlag:        true,
			breakPoint:       make([]uint64, 0),
			codeBreakPointed: make([]bool, code.CodeCount),
		}
	} else {
		ret = &CodeRunner{
			code:   code,
			memory: memory.New(),
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

// PrintBreakPoint prints all breakpoints and watching information.
func (cr *CodeRunner) PrintDebugInfo() {
	if !cr.debugFlag {
		panic("CodeRunner: can't print breakpoints when not in debug mode")
	}

	// Print breakpoint info
	if len(cr.breakPoint) == 0 {
		fmt.Print("No breakpoints exist now.\n\n")
	} else {
		// Print each breakpoint
		fmt.Println("Num\tLine")
		for index, line := range cr.breakPoint {
			fmt.Printf("%v\t%v\n", index+1, line)
		}
		fmt.Print("\n")
	}

	// Print watching info
	if cr.isWatching {
		fmt.Printf("Watching memory at offset %v\n\n", cr.watchOffset)
	} else {
		fmt.Print("No memory watching now.\n\n")
	}
}

// PrintCode prints analysed code information.
func (cr *CodeRunner) PrintAllOperator() {
	cr.code.PrintAll()
}

// PrintNextOperator prints the next operator to be executed.
func (cr *CodeRunner) PrintNextOperator() {
	if cr.codeIndex >= cr.code.CodeCount {
		cr.code.Print(0)
	} else {
		cr.code.Print(cr.codeIndex)
	}
}

// PeekBytes peeks bytes from memory with the given offset and length.
//
// Offset is relative to the current memory pointer.
func (cr *CodeRunner) PeekBytes(offset, length int) (ret []byte) {
	return cr.memory.PeekBytes(offset, length)
}

// Watch sets a watch on the memory byte at the current pointer plus the given offset.
func (cr *CodeRunner) Watch(offset int) {
	cr.isWatching = true
	cr.watchOffset = offset
	fmt.Printf("Watch at offset %v\n\n", offset)
}

// UntilLoopEnd runs the code until the current loop (enclosed by []) ends.
func (cr *CodeRunner) UntilLoopEnd() {
	if cr.untilStatus {
		fmt.Print("Already in until mode\n\n")
		return
	} else {
		cr.untilStatus = true
		fmt.Print("Entering until mode\n\n")
	}
}

// Run starts running the code from the beginning.
func (cr *CodeRunner) Run() (ret ReturnCode) {
	cr.Reset()
	ret = cr.Continue()
	return
}

// Continue continues running the code from the current position.
func (cr *CodeRunner) Continue() (ret ReturnCode) {
	for {
		// Check for breakpoint
		if cr.breakPointUsed {
			cr.breakPointUsed = false
		} else if cr.debugFlag && cr.codeBreakPointed[cr.codeIndex] {
			// Hit breakpoint
			cr.breakPointUsed = true
			return ReturnReachBreakPoint
		}

		// Execute operator
		if ret = cr.executeOperator(); ret != returnAfterExecuteOperator {
			if ret == ReturnAfterFinish {
				cr.clearDebugFlags()
			}
			return ret
		}
	}
}

// Step executes the next operator, ignore breakpoints.
func (cr *CodeRunner) Step() (ret ReturnCode) {
	// Check finish, reset if finished
	if cr.codeIndex >= cr.code.CodeCount {
		cr.Reset()
	}

	// Execute operator
	ret = cr.executeOperator()

	// Check finish
	switch ret {
	case returnAfterExecuteOperator:
		ret = ReturnAfterStep

	case ReturnAfterFinish:
		cr.clearDebugFlags()
	}
	return
}

// executeOperator executes the current operator and advances the code index.
func (cr *CodeRunner) executeOperator() (ret ReturnCode) {
	var operator code.Operator = cr.code.Operators[cr.codeIndex]
	var auxiliary uint64 = cr.code.Auxiliary[cr.codeIndex]
	cr.codeIndex++

	switch operator {
	case code.OpAdd:
		if cr.isWatching && cr.watchOffset == 0 {
			cr.isWatching = false
			cr.codeIndex--
			return ReturnReachWatch
		}
		cr.memory.Add(auxiliary)

	case code.OpSub:
		if cr.isWatching && cr.watchOffset == 0 {
			cr.isWatching = false
			cr.codeIndex--
			return ReturnReachWatch
		}
		cr.memory.Sub(auxiliary)

	case code.OpMoveLeft:
		// Memory block may change after moving pointer
		cr.memory = cr.memory.MovePtr(-int(auxiliary))
		cr.watchOffset += int(auxiliary)

	case code.OpMoveRight:
		// Memory block may change after moving pointer
		cr.memory = cr.memory.MovePtr(int(auxiliary))
		cr.watchOffset -= int(auxiliary)

	case code.OpLeftBracket:
		if cr.memory.Peek(0) == 0 {
			cr.codeIndex = int(auxiliary)
		}

	case code.OpRightBracket:
		if cr.memory.Peek(0) != 0 {
			cr.codeIndex = int(auxiliary)
			// Check until mode, if running has finished, ignore until
			if cr.untilStatus && cr.codeIndex < cr.code.CodeCount {
				cr.untilStatus = false
				return ReturnReachUntil
			}
		}

	case code.OpInput:
		var input rune
		fmt.Scanf("%c", &input)
		cr.memory.Poke(byte(input))

	case code.OpOutput:
		fmt.Printf("%c", cr.memory.Peek(0))
	}

	if cr.codeIndex >= cr.code.CodeCount {
		return ReturnAfterFinish
	} else {
		return returnAfterExecuteOperator
	}
}

// Reset resets the CodeRunner to the initial state.
func (cr *CodeRunner) Reset() {
	cr.codeIndex = 0
	cr.memory = memory.New()
}

// clearDebugFlags clears all debug-related flags.
func (cr *CodeRunner) clearDebugFlags() {
	cr.isWatching = false
	cr.untilStatus = false
}
