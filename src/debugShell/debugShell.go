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
	"s[tep]:				  : Step to next operator \n" +
	"d[etailed] [times]     : Detailed step for specified times, show each operator and tape after execute, default run until finish\n" +
	"u[ntil]                  : Run until loop([]) finish\n" +
	"w[atch] <offset>         : Watch memory byte at current pointer plus offset\n" +
	"b[reak] <line>           : Set breakpoint at specified line\n" +
	"del[ete] <line>          : Delete breakpoint at specified line\n" +
	"i[nfo]                   : Information of breakpoints and watching\n" +
	"clear                    : Clear all breakpoints\n" +
	"p[eek] [offset [length]] : Peek memory bytes at current pointer with optional offset and length\n" +
	"t[ape]                   : Show tape around, equal to peek -10 20\n" +
	"n[ext]                   : Show next operator to be executed\n" +
	"code                     : Show analysed code information\n" +
	"h[elp]                   : Show this help message\n" +
	"q[uit]                   : Quit debug shell\n" +
	"\n"

var REG_DETAILED *regexp.Regexp = regexp.MustCompile(`^d(etailed)?( (\d+))?$`)
var REG_WATCH *regexp.Regexp = regexp.MustCompile(`^w(atch)? (-?\d+)$`)
var REG_BREAK *regexp.Regexp = regexp.MustCompile(`^b(reak)? (\d+)$`)
var REG_DELETE *regexp.Regexp = regexp.MustCompile(`^del(ete)? (\d+)$`)
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

		case "s", "step":
			var ret coderunner.ReturnCode = codeRunner.Step()

			// Check return code
			switch ret {
			case coderunner.ReturnReachWatch:
				fmt.Print("Watch hit\n\n")
				CodeRunning = true

			case coderunner.ReturnReachUntil:
				fmt.Print("Until finished\n\n")
				CodeRunning = true

			case coderunner.ReturnAfterFinish:
				fmt.Print("\n\nRunning finished\n\n")
				CodeRunning = false

			case coderunner.ReturnAfterStep:
				fmt.Print("\n")
				CodeRunning = true

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

		case "i", "info":
			codeRunner.PrintDebugInfo()
			continue

		case "clear":
			codeRunner.ClearBreakPoints()
			continue

		case "n", "next":
			codeRunner.PrintNextOperator()
			fmt.Print("\n") // Extra newline for better readability
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

		if matches := REG_WATCH.FindStringSubmatch(command); matches != nil {
			// regex match watch command
			var offset int
			fmt.Sscanf(matches[2], "%d", &offset)
			codeRunner.Watch(offset)
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
			var index int
			fmt.Sscanf(matches[2], "%d", &index)

			err := codeRunner.RemoveBreakPoint(index)
			if err != nil {
				fmt.Printf("%s\n\n", err.Error())
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

// detailedStep performs a single step and prints detailed information.
func detailedStep(codeRunner *coderunner.CodeRunner) (ret coderunner.ReturnCode) {
	codeRunner.PrintNextOperator()
	fmt.Print("\n")
	ret = codeRunner.Step()
	// Print tape around
	peekTape(codeRunner, -10, 20)
	return
}
