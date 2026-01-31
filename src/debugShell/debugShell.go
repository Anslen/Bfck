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

package debugshell

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	coderunner "github.com/Anslen/Bfck/codeManager/codeRunner"
)

const HELP_STRING string = "Execute commands:\n" +
	"r[un]                    : Run code from begin\n" +
	"c[ontinue]               : Continue run code\n" +
	"s[tep] [times]           : Step by times, default 1\n" +
	"d[etailed] [times]       : Detailed step for specified times, default run until finish\n" +
	"u[ntil]                  : Run until loop([]) finish\n" +
	"\nDebug commands:\n" +
	"b[reak] <line>           : Set breakpoint at specified line\n" +
	"w[atch] <address>        : Watch memory at address\n" +
	"del[ete] b|w <num>       : Delete breakpoint or watchpoint at specified number\n" +
	"i[nfo] [b|w]             : Information of breakpoints or watching, default both\n" +
	"clear [b|w]              : Clear all breakpoints or watchpoints, default both\n" +
	"\nMemory commands:\n" +
	"ptr                      : Show current memory pointer\n" +
	"p[eek] [offset [length]] : Peek memory bytes at current pointer with optional offset and length\n" +
	"t[ape]                   : Show tape around, equal to peek -10 20\n" +
	"reset                    : Reset memory tape immediately\n" +
	"\nOther commands:\n" +
	"n[ext]                   : Show next operator to be executed\n" +
	"code                     : Show analysed code information\n" +
	"h[elp]                   : Show this help message\n" +
	"q[uit]                   : Quit debug shell\n" +
	"\n"

var REG_STEP *regexp.Regexp = regexp.MustCompile(`^s(tep)?( (\d+))?$`)
var REG_DETAILED *regexp.Regexp = regexp.MustCompile(`^d(etailed)?( (\d+))?$`)
var REG_WATCH *regexp.Regexp = regexp.MustCompile(`^w(atch)? (-?\d+)$`)
var REG_BREAK *regexp.Regexp = regexp.MustCompile(`^b(reak)? (\d+)$`)
var REG_DELETE *regexp.Regexp = regexp.MustCompile(`^del(ete)? (b|w) (\d+)$`)
var REG_INFO *regexp.Regexp = regexp.MustCompile(`^i(nfo)?( (b|w))?$`)
var REG_CLEAR *regexp.Regexp = regexp.MustCompile(`^clear( (b|w))?$`)
var REG_PEEK *regexp.Regexp = regexp.MustCompile(`^p(eek)?( (-?\d+)( (\d+))?)?$`)

var DEBUG_REG_FUNCTIONS = []func(string, *coderunner.CodeRunner) bool{
	regMatchBreak,
	regMatchWatch,
	regMatchDelete,
	regMatchInfo,
	regMatchClear,
	regMatchPeek,
}

// Start starts the debug shell for the given code runner.
func Start(codeRunner *coderunner.CodeRunner) {
	var CodeRunning bool = false
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("(Bfck) ")

		// Read command
		if !scanner.Scan() {
			break
		}
		command := strings.TrimSpace(scanner.Text())
		if command == "" {
			continue
		}

		// Quit command
		if command == "q" || command == "quit" {
			break
		}

		// Match simple commands
		if matchSimpleCommands(command, codeRunner, &CodeRunning) {
			continue
		}

		// Match regex commands
		if matchRegexCommands(command, codeRunner, &CodeRunning) {
			continue
		}

		// No match command
		fmt.Print("Unknown command. Type h for help\n\n")
	}
}

// matchSimpleCommands matches simple commands that does not require regex.
func matchSimpleCommands(command string, codeRunner *coderunner.CodeRunner, codeRunning *bool) bool {
	switch command {
	case "r", "run":
		// Run code from beginning and get return code
		printDebugMessage(codeRunner.Run(), codeRunning)
		return true

	case "c", "continue":
		// Check if code is running
		if !*codeRunning {
			fmt.Print("Code is not running. Use 'run' command to start.\n\n")
			return true
		}

		// Continue running code
		printDebugMessage(codeRunner.Continue(), codeRunning)
		return true

	case "u", "until":
		// Check if code is running
		if !*codeRunning {
			fmt.Print("Code is not running. Use 'run' command to start.\n\n")
		} else {
			codeRunner.EnableUntil()
		}
		return true

	case "ptr":
		var ptr int = codeRunner.GetMemoryPointer()
		fmt.Printf("Current memory pointer: %d\n\n", ptr)
		return true

	case "t", "tape":
		// Print memory pointer
		var ptr int = codeRunner.GetMemoryPointer()
		fmt.Printf("Current memory pointer: %d\n", ptr)

		// Peek tape around
		peekTape(codeRunner, -10, 20)

		return true

	case "reset":
		codeRunner.Reset()
		fmt.Print("Memory tape reseted.\n\n")
		return true

	case "n", "next":
		codeRunner.PrintNextOperator()
		fmt.Print("\n") // Extra newline for better readability
		return true

	case "code":
		codeRunner.PrintAllOperators()
		return true

	case "h", "help":
		fmt.Print(HELP_STRING)
		return true
	}
	return false
}

// matchRegexCommands tries to match the command with regex commands.
func matchRegexCommands(command string, codeRunner *coderunner.CodeRunner, codeRunning *bool) bool {
	if regMatchStep(command, codeRunner, codeRunning) {
		return true
	}
	if regMatchDetailed(command, codeRunner, codeRunning) {
		return true
	}

	for _, function := range DEBUG_REG_FUNCTIONS {
		if function(command, codeRunner) {
			return true
		}
	}

	return false
}

// regMatchStep regex matching and executing step command.
func regMatchStep(command string, codeRunner *coderunner.CodeRunner, codeRunning *bool) bool {
	// Match regex
	var matches []string = REG_STEP.FindStringSubmatch(command)
	if matches == nil {
		return false
	}

	// Read arguments
	var times int
	if matches[3] == "" {
		times = 1
	} else {
		fmt.Sscanf(matches[3], "%d", &times)
	}

	// Execute step
	for i := 0; i < times; i++ {
		var ret coderunner.ReturnCode = step(codeRunner, codeRunning)
		if ret == coderunner.ReturnAfterFinish {
			break
		}
	}
	fmt.Print("\n")
	return true
}

// regMatchDetailed regex matching and executing detailed command.
func regMatchDetailed(command string, codeRunner *coderunner.CodeRunner, codeRunning *bool) bool {
	// Match regex
	var matches []string = REG_DETAILED.FindStringSubmatch(command)
	if matches == nil {
		return false
	}

	// Read arguments
	var times uint64
	if matches[3] == "" {
		times = ^uint64(0)
	} else {
		fmt.Sscanf(matches[3], "%d", &times)
	}

	// Execute detailed step
	var i uint64
	for i = 0; i < times; i++ {
		var ret coderunner.ReturnCode = detailedStep(codeRunner, codeRunning)
		// Break when finished
		if ret == coderunner.ReturnAfterFinish {
			break
		}
	}
	return true
}

// regMatchBreak regex matching and executing break command.
func regMatchBreak(command string, codeRunner *coderunner.CodeRunner) bool {
	// Match regex
	var matches []string = REG_BREAK.FindStringSubmatch(command)
	if matches == nil {
		return false
	}

	// Read arguments
	var line uint64
	fmt.Sscanf(matches[2], "%d", &line)

	// Execute break
	var message string = codeRunner.AddBreakPoint(line)
	fmt.Print(message)
	return true
}

// regMatchWatch regex matching and executing watch command.
func regMatchWatch(command string, codeRunner *coderunner.CodeRunner) bool {
	// Match regex
	var matches []string = REG_WATCH.FindStringSubmatch(command)
	if matches == nil {
		return false
	}

	// Read arguments
	var address int
	fmt.Sscanf(matches[2], "%d", &address)

	// Execute watch
	var message string = codeRunner.AddWatch(address)
	fmt.Print(message)
	return true
}

// regMatchDelete regex matching and executing delete command.
func regMatchDelete(command string, codeRunner *coderunner.CodeRunner) bool {
	// Match regex
	var matches []string = REG_DELETE.FindStringSubmatch(command)
	if matches == nil {
		return false
	}

	// Read index
	var index int
	fmt.Sscanf(matches[3], "%d", &index)

	// Remove according to type
	var message string
	switch matches[2] {
	case "b":
		message = codeRunner.RemoveBreakPoint(index)

	case "w":
		message = codeRunner.RemoveWatch(index)

	default:
		panic("DebugShell: Invalid delete command")
	}

	// Print result message
	fmt.Print(message)
	return true
}

// regMatchInfo regex matching and executing info command.
func regMatchInfo(command string, codeRunner *coderunner.CodeRunner) bool {
	// Match regex
	var matches []string = REG_INFO.FindStringSubmatch(command)
	if matches == nil {
		return false
	}

	// Execute info
	switch matches[3] {
	case "b":
		codeRunner.PrintBreakPoints()

	case "w":
		codeRunner.PrintWatchInfo()

	case "":
		codeRunner.PrintBreakPoints()
		codeRunner.PrintWatchInfo()

	default:
		panic("DebugShell: Invalid info command")
	}
	return true
}

// regMatchClear regex matching and executing clear command.
func regMatchClear(command string, codeRunner *coderunner.CodeRunner) bool {
	// Match regex
	var matches []string = REG_CLEAR.FindStringSubmatch(command)
	if matches == nil {
		return false
	}

	// Execute clear
	switch matches[2] {
	case "b":
		codeRunner.ClearBreakPoints()
		fmt.Print("All breakpoints cleared\n\n")

	case "w":
		codeRunner.ClearWatches()
		fmt.Print("All watchpoints cleared\n\n")

	case "":
		codeRunner.ClearBreakPoints()
		codeRunner.ClearWatches()
		fmt.Print("All breakpoints and watchpoints cleared\n\n")

	default:
		panic("DebugShell: Invalid clear command")
	}
	return true
}

// regMatchPeek regex matching and executing peek command.
func regMatchPeek(command string, codeRunner *coderunner.CodeRunner) bool {
	// Match regex
	var matches []string = REG_PEEK.FindStringSubmatch(command)
	if matches == nil {
		return false
	}

	// Read arguments
	var offset, length int
	// Read offset
	if matches[3] == "" {
		offset = 0
	} else {
		fmt.Sscanf(matches[3], "%d", &offset)
	}
	// Read length
	if matches[5] == "" {
		length = 1
	} else {
		fmt.Sscanf(matches[5], "%d", &length)
	}

	// Execute peek
	peekTape(codeRunner, offset, length)
	return true
}

// printDebugMessage prints debug messages according to the return code.
//
// Used in run and continue commands.
func printDebugMessage(ret coderunner.ReturnCode, codeRunning *bool) {
	switch ret {
	case coderunner.ReturnReachBreakPoint:
		fmt.Print("\n\nHit breakpoint\n\n")
		*codeRunning = true

	case coderunner.ReturnReachWatch:
		fmt.Print("\n\nWatch hit\n\n")
		*codeRunning = true

	case coderunner.ReturnReachUntil:
		fmt.Print("\n\nUntil finished\n\n")
		*codeRunning = true

	case coderunner.ReturnAfterFinish:
		fmt.Print("\n\nRunning finished\n\n")
		*codeRunning = false

	default:
		panic("DebugShell: Unknown return code")
	}
}

// peekTape peeks memory bytes at the given offset and length, and prints them.
func peekTape(codeRunner *coderunner.CodeRunner, offset, length int) {
	var bytes []byte = codeRunner.PeekBytes(offset, length)
	// Print bytes
	for index, each := range bytes {
		if offset+index == 0 {
			fmt.Printf("[%d] ", each)
		} else {
			fmt.Printf("%d ", each)
		}
	}
	fmt.Print("\n\n")
}

// step performs a single step and updates the code running status.
//
// Return message is displayed in this function.
func step(codeRunner *coderunner.CodeRunner, codeRunning *bool) (ret coderunner.ReturnCode) {
	ret = codeRunner.Step()

	// Check return code
	// Step show message briefly so don't use checkReturnCode function
	switch ret {
	case coderunner.ReturnReachWatch:
		fmt.Print("Watch hit\n\n")
		*codeRunning = true

	case coderunner.ReturnReachUntil:
		fmt.Print("Until finished\n\n")
		*codeRunning = true

	case coderunner.ReturnAfterFinish:
		fmt.Print("\n\nRunning finished\n\n")
		*codeRunning = false

	case coderunner.ReturnAfterStep:
		*codeRunning = true

	default:
		panic("DebugShell: Invalid return code")
	}

	return
}

// detailedStep performs a single step and prints detailed information.
//
// Return message is displayed in this function.
func detailedStep(codeRunner *coderunner.CodeRunner, codeRunning *bool) (ret coderunner.ReturnCode) {
	// Show next operator
	codeRunner.PrintNextOperator()
	fmt.Print("\n")

	// Get and print memory pointer
	var memoryPointer int = codeRunner.GetMemoryPointer()
	fmt.Printf("Memory pointer at: %d\n", memoryPointer)

	// Step code
	ret = codeRunner.Step()

	// Print tape around
	peekTape(codeRunner, -10, 20)

	// Check return code
	if ret == coderunner.ReturnAfterFinish {
		fmt.Print("\n\nRunning finished\n\n")
		*codeRunning = false
	} else {
		*codeRunning = true
	}

	return
}
