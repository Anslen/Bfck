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

const HELP_STRING string = "r[un]                    : Run code from begin\n" +
	"c[ontinue]               : Continue running code until next breakpoint or end\n" +
	"s[tep] [times]           : Step by times, default 1\n" +
	"d[etailed] [times]       : Detailed step for specified times, default run until finish\n" +
	"u[ntil]                  : Run until loop([]) finish\n" +
	"w[atch] <address>        : Watch memory at address\n" +
	"b[reak] <line>           : Set breakpoint at specified line\n" +
	"del[ete] b|w <num>       : Delete breakpoint or watchpoint at specified number\n" +
	"i[nfo] [b|w]             : Information of breakpoints or watching, default both\n" +
	"clear [b|w]              : Clear all breakpoints or watchpoints, default both\n" +
	"ptr                      : Show current memory pointer\n" +
	"p[eek] [offset [length]] : Peek memory bytes at current pointer with optional offset and length\n" +
	"t[ape]                   : Show tape around, equal to peek -10 20\n" +
	"n[ext]                   : Show next operator to be executed\n" +
	"reset                    : Reset memory tape immediately\n" +
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

// Start starts the debug shell for the given code runner.
func Start(codeRunner *coderunner.CodeRunner) {
	var CodeRunning bool = false
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("(Bfck) ")
		if !scanner.Scan() {
			break
		}
		command := strings.TrimSpace(scanner.Text())
		if command == "" {
			continue
		}

		switch command {
		case "r", "run":
			// Run code from beginning and get return code
			var ret coderunner.ReturnCode = codeRunner.Run()
			switch ret {
			case coderunner.ReturnReachBreakPoint:
				fmt.Print("\n\nHit breakpoint\n\n")
				CodeRunning = true

			case coderunner.ReturnReachWatch:
				fmt.Print("\n\nWatch hit\n\n")
				CodeRunning = true

			case coderunner.ReturnReachUntil:
				fmt.Print("\n\nUntil finished\n\n")
				CodeRunning = true

			case coderunner.ReturnAfterFinish:
				fmt.Print("\n\nRunning finished\n\n")
				CodeRunning = false

			default:
				panic("DebugShell: Unknown return code")
			}
			continue

		case "c", "continue":
			// Check if code is running
			if !CodeRunning {
				fmt.Print("Code is not running. Use 'run' command to start.\n\n")
				continue
			}

			// Continue running code
			var ret coderunner.ReturnCode = codeRunner.Continue()
			switch ret {
			case coderunner.ReturnReachBreakPoint:
				fmt.Print("\n\nHit breakpoint\n\n")
				CodeRunning = true

			case coderunner.ReturnReachWatch:
				fmt.Print("\n\nWatch hit\n\n")
				CodeRunning = true

			case coderunner.ReturnReachUntil:
				fmt.Print("\n\nUntil finished\n\n")
				CodeRunning = true

			case coderunner.ReturnAfterFinish:
				fmt.Print("\n\nRunning finished\n\n")
				CodeRunning = false

			default:
				panic("DebugShell: Invalid return code")
			}
			continue

		case "t", "tape":
			peekTape(codeRunner, -10, 20)
			continue

		case "u", "until":
			// Check if code is running
			if !CodeRunning {
				fmt.Print("Code is not running. Use 'run' command to start.\n\n")
			} else {
				codeRunner.UntilLoopEnd()
			}
			continue

		case "ptr":
			var ptr int = codeRunner.GetMemoryPointer()
			fmt.Printf("Current memory pointer: %d\n\n", ptr)
			continue

		case "n", "next":
			codeRunner.PrintNextOperator()
			fmt.Print("\n") // Extra newline for better readability
			continue

		case "reset":
			codeRunner.Reset()
			fmt.Print("Memory tape reseted.\n\n")
			continue

		case "code":
			codeRunner.PrintAllOperator()
			continue

		case "h", "help":
			fmt.Print(HELP_STRING)
			continue

		case "q", "quit":
			return
		}

		if matches := REG_DETAILED.FindStringSubmatch(command); matches != nil {
			// regex match detailed command
			var times uint64
			if matches[3] == "" {
				times = ^uint64(0)
			} else {
				fmt.Sscanf(matches[3], "%d", &times)
			}

			var i uint64
			for i = 0; i < times; i++ {
				var ret coderunner.ReturnCode = detailedStep(codeRunner)
				// Check return code
				if ret == coderunner.ReturnAfterFinish {
					fmt.Print("\n\nRunning finished\n\n")
					CodeRunning = false
					break
				} else {
					CodeRunning = true
				}
			}
			continue
		}

		if matches := REG_STEP.FindStringSubmatch(command); matches != nil {
			// regex match step command
			var times int
			if matches[3] == "" {
				times = 1
			} else {
				fmt.Sscanf(matches[3], "%d", &times)
			}

			for i := 0; i < times; i++ {
				var ret coderunner.ReturnCode = step(codeRunner, &CodeRunning)
				if ret == coderunner.ReturnAfterFinish {
					break
				}
			}
			fmt.Print("\n")
			continue
		}

		if matches := REG_WATCH.FindStringSubmatch(command); matches != nil {
			// regex match watch command
			var address int
			fmt.Sscanf(matches[2], "%d", &address)
			var message string = codeRunner.AddWatch(address)
			fmt.Print(message)
			continue
		}

		if matches := REG_BREAK.FindStringSubmatch(command); matches != nil {
			// regex match break command
			var line uint64
			fmt.Sscanf(matches[2], "%d", &line)
			err := codeRunner.AddBreakPoint(line)

			if err != nil {
				fmt.Printf("%s\n\n", err.Error())
			}
			continue
		}

		if matches := REG_DELETE.FindStringSubmatch(command); matches != nil {
			// regex match delete command
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
			continue
		}

		if matches := REG_INFO.FindStringSubmatch(command); matches != nil {
			// regex match info command
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
			continue
		}

		if matches := REG_CLEAR.FindStringSubmatch(command); matches != nil {
			// regex match clear command
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
			continue
		}

		if matches := REG_PEEK.FindStringSubmatch(command); matches != nil {
			// regex match peek command
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

			peekTape(codeRunner, offset, length)
			continue
		}

		// No match command, print help
		fmt.Print("Unknown command. Type h for help\n\n")
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

func step(codeRunner *coderunner.CodeRunner, codeRunning *bool) (ret coderunner.ReturnCode) {
	ret = codeRunner.Step()

	// Check return code
	switch ret {
	case coderunner.ReturnReachWatch:
		fmt.Print("Watch hit\n\n")
		*codeRunning = true

	case coderunner.ReturnReachUntil:
		fmt.Print("Until finished\n\n")
		*codeRunning = true

	case coderunner.ReturnAfterFinish:
		fmt.Print("\n\nRunning finished\n")
		*codeRunning = false

	case coderunner.ReturnAfterStep:
		*codeRunning = true

	default:
		panic("DebugShell: Invalid return code")
	}

	return
}

// detailedStep performs a single step and prints detailed information.
func detailedStep(codeRunner *coderunner.CodeRunner) (ret coderunner.ReturnCode) {
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
	return
}
