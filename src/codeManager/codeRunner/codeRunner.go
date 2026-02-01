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
	ReturnReachStop
	returnAfterExecuteOperator // For internal function executeOperator
)

type CodeRunner struct {
	code             *code.Code
	codeIndex        int // Point at next operator to execute
	memory           *memory.Memory
	memoryPointer    int
	debugFlag        bool
	breakPoint       []uint64
	codeBreakPointed []bool
	breakPointUsed   bool
	watchAddress     []int
	watchUsed        bool
	watchChecked     bool
	watchHit         bool
	untilStatus      bool
	stopEnabled      bool
	stopIndex        int
	untilEnabled     bool
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
			watchAddress:     make([]int, 0),
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
func (cr *CodeRunner) AddBreakPoint(line uint64) (message string) {
	if !cr.debugFlag {
		panic("CodeRunner: can't add breakpoint when not in debug mode")
	}

	// Check line range
	if line == 0 || line > cr.code.LineCount {
		message = fmt.Sprintf("Error: breakpoint out of range, line count is %v, get line %v", cr.code.LineCount, line)
		return
	}

	// Find position and add breakpoint
	position, found := slices.BinarySearch(cr.breakPoint, line)
	if !found {
		// Warn if breakpoint at empty line
		if cr.code.LineBegins[line-1] == -1 {
			message = fmt.Sprintf("Warning: breakpoint at line %v will not work\n\n", line)
			return
		}

		cr.codeBreakPointed[cr.code.LineBegins[line-1]] = true
		cr.breakPoint = slices.Insert(cr.breakPoint, position, line)
		// Pirnt add success information
		message = fmt.Sprintf("Breakpoint added at line %v\n\n", line)

	} else {
		message = fmt.Sprintf("Warning: Breakpoint at line %v already existed\n\n", line)
	}

	return
}

// RemoveBreakPoint removes the breakpoint at the specified index.
//
// CAUSION: index start from 1
func (cr *CodeRunner) RemoveBreakPoint(index int) (message string) {
	if !cr.debugFlag {
		panic("CodeRunner: can't remove breakpoint when not in debug mode")
	}

	// Check index range
	if index <= 0 || index > len(cr.breakPoint) {
		message = fmt.Sprintf("Error: breakpoint index out of range, get %v, breakpoint count is %v", index, len(cr.breakPoint))
		return
	}

	// Remove success information
	message = fmt.Sprintf("Breakpoint %v removed\n\n", index)

	// Remove breakpoint
	var removedLine uint64 = cr.breakPoint[index-1]
	var removedCodeIndex int = cr.code.LineBegins[removedLine-1]
	cr.breakPoint = slices.Delete(cr.breakPoint, index-1, index)

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
	cr.codeBreakPointed = make([]bool, cr.code.CodeCount)
}

// PrintBreakPoint prints all breakpoints and watching information.
func (cr *CodeRunner) PrintBreakPoints() {
	if !cr.debugFlag {
		panic("CodeRunner: can't print breakpoints when not in debug mode")
	}

	// Print breakpoint info
	if len(cr.breakPoint) == 0 {
		fmt.Print("No breakpoints exist now.\n\n")
	} else {
		// Print each breakpoint
		fmt.Println("Breakpoints:")
		fmt.Println("Num\tLine")
		for index, line := range cr.breakPoint {
			fmt.Printf("%v\t%v\n", index+1, line)
		}
		fmt.Print("\n")
	}
}

// AddWatch sets a watch on the memory byte at the current pointer plus the given offset.
func (cr *CodeRunner) AddWatch(address int) (message string) {
	var (
		position int
		found    bool
	)
	position, found = slices.BinarySearch(cr.watchAddress, address)
	if found {
		message = fmt.Sprintf("Warning: Address %v is already being watched\n\n", address)
	} else {
		cr.watchAddress = slices.Insert(cr.watchAddress, position, address)
		message = fmt.Sprintf("Watching memory %v\n\n", address)
	}
	return
}

func (cr *CodeRunner) RemoveWatch(index int) (message string) {
	if index <= 0 || index > len(cr.watchAddress) {
		message = fmt.Sprintf("Error: Watchpoint index out of range, get %v, watchpoint count is %v\n\n", index, len(cr.watchAddress))
		return
	}

	message = fmt.Sprintf("Watchpoint %v at address %v removed\n\n", index, cr.watchAddress[index-1])
	cr.watchAddress = slices.Delete(cr.watchAddress, index-1, index)
	return
}

func (cr *CodeRunner) ClearWatches() {
	if !cr.debugFlag {
		panic("CodeRunner: can't clear watchpoints when not in debug mode")
	}

	cr.watchAddress = make([]int, 0)
}

// PrintWatchInfo prints all watchpoints information.
func (cr *CodeRunner) PrintWatchInfo() {
	if !cr.debugFlag {
		panic("CodeRunner: can't print watchpoints when not in debug mode")
	}

	// Print watch info
	if len(cr.watchAddress) == 0 {
		fmt.Print("No watchpoints exist now.\n\n")
	} else {
		// Print each watchpoint
		fmt.Println("Watch address:")
		fmt.Println("Num\tAddress")
		for index, address := range cr.watchAddress {
			fmt.Printf("%v\t%v\n", index+1, address)
		}
		fmt.Print("\n")
	}
}

// SetStopPoint sets the code runner to stop execution at the specified operator index.
func (cr *CodeRunner) SetStopPoint(index int) {
	cr.stopEnabled = true
	cr.stopIndex = index
}

// RemoveStopPoint removes the stop point.
func (cr *CodeRunner) RemoveStopPoint() (message string) {
	if cr.stopEnabled {
		cr.stopEnabled = false
		message = "Stop point removed\n\n"
	} else {
		message = "No stop point setted now\n\n"

	}
	return
}

// PrintStopPoint prints the current stop point information.
func (cr *CodeRunner) PrintStopPoint() {
	if cr.stopEnabled {
		fmt.Printf("Stop point set at operator index %v\n\n", cr.stopIndex)
	} else {
		fmt.Print("No stop point setted now.\n\n")
	}
}

func (cr *CodeRunner) PrintAllDebugInfo() {
	cr.PrintBreakPoints()
	cr.PrintWatchInfo()
	cr.PrintStopPoint()
}

// PrintCode prints analysed code information.
func (cr *CodeRunner) PrintAllOperators() {
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

// GetMemoryPointer returns the current memory pointer.
func (cr *CodeRunner) GetMemoryPointer() int {
	return cr.memoryPointer
}

// PeekBytes peeks bytes from memory with the given offset and length.
//
// Offset is relative to the current memory pointer.
func (cr *CodeRunner) PeekBytes(offset, length int) (ret []byte) {
	return cr.memory.PeekBytes(offset, length)
}

// EnableUntil enables the until mode.
func (cr *CodeRunner) EnableUntil() {
	if cr.untilEnabled {
		fmt.Print("Already in until mode\n\n")
		return
	} else {
		cr.untilEnabled = true
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

	// Change return code if after execute operator
	if ret == returnAfterExecuteOperator {
		ret = ReturnAfterStep
	}

	return
}

// executeOperator executes the current operator and advances the code index.
func (cr *CodeRunner) executeOperator() (ret ReturnCode) {
	// Check stop point
	if cr.debugFlag && cr.stopEnabled && cr.codeIndex == cr.stopIndex {
		cr.stopEnabled = false
		return ReturnReachStop
	}

	// fetch operator and auxiliary data
	var operator code.Operator = cr.code.Operators[cr.codeIndex]
	var auxiliary uint64 = cr.code.Auxiliary[cr.codeIndex]
	cr.codeIndex++

	// Execute operator
	switch operator {
	case code.OpAdd:
		// Check watchpoint
		if cr.debugFlag && cr.isWatchHit() {
			cr.codeIndex--
			return ReturnReachWatch
		}

		// Execute addition
		cr.memory.Add(auxiliary)
		cr.watchUsed = false

	case code.OpSub:
		// Check watchpoint
		if cr.debugFlag && cr.isWatchHit() {
			cr.codeIndex--
			return ReturnReachWatch
		}

		// Execute subtraction
		cr.memory.Sub(auxiliary)
		cr.watchUsed = false

	case code.OpMoveLeft:
		// Memory block may change after moving pointer
		cr.memory = cr.memory.MovePtr(-int(auxiliary))
		cr.memoryPointer -= int(auxiliary)
		cr.watchChecked = false

	case code.OpMoveRight:
		// Memory block may change after moving pointer
		cr.memory = cr.memory.MovePtr(int(auxiliary))
		cr.memoryPointer += int(auxiliary)
		cr.watchChecked = false

	case code.OpLeftBracket:
		if cr.memory.Peek(0) == 0 {
			cr.codeIndex = int(auxiliary)
		}

	case code.OpRightBracket:
		if cr.memory.Peek(0) != 0 {
			cr.codeIndex = int(auxiliary)
		} else if cr.untilEnabled {
			// Check until mode
			cr.untilEnabled = false
			return ReturnReachUntil
		}

	case code.OpInput:
		if cr.debugFlag && cr.isWatchHit() {
			cr.codeIndex--
			return ReturnReachWatch
		}

		var input rune
		fmt.Scanf("%c", &input)
		cr.memory.Poke(byte(input))
		cr.watchUsed = false

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
	// Reset code index and memory
	cr.codeIndex = 0
	cr.memory = memory.New()
	cr.memoryPointer = 0

	// Clear debug flags
	cr.breakPointUsed = false
	cr.watchUsed = false
	cr.untilEnabled = false
}

// isWatchHit checks if the current memory pointer hits any watchpoint.
func (cr *CodeRunner) isWatchHit() bool {
	if !cr.debugFlag {
		panic("CodeRunner: can't check watch hit when not in debug mode")
	}

	// Check watchpoint only once after memory pointer changes
	if !cr.watchChecked {
		cr.watchChecked = true
		var found bool
		_, found = slices.BinarySearch(cr.watchAddress, cr.memoryPointer)
		if found {
			cr.watchHit = true
		}
	}

	// Return watch hit status, if watch used, flip the status and continue
	if cr.watchHit {
		cr.watchUsed = !cr.watchUsed
		return cr.watchUsed
	} else {
		return false
	}
}
