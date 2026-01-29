package debugshell

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	coderunner "github.com/Anslen/Bfck/codeManager/codeRunner"
)

const HELP_STRING string = "r[un]           : Run code from begin\n" +
	"c[ontinue]      : Continue running code until next breakpoint or end\n" +
	"b[reak] <line>  : Set breakpoint at specified line\n" +
	"d[elete] <line> : Delete breakpoint at specified line\n" +
	"l[ist]          : List all breakpoints\n" +
	"s[ee]           : See analysed code information\n" +
	"h[elp]          : Show this help message\n" +
	"q[uit]          : Quit debug shell\n" +
	"\n"

var REG_BREAK *regexp.Regexp = regexp.MustCompile(`^b(reak)? (\d+)$`)
var REG_DELETE *regexp.Regexp = regexp.MustCompile(`^d(elete)? (\d+)$`)
var REG_PEEK *regexp.Regexp = regexp.MustCompile(`^p(eek)?( (-?\d+)( (\d+))?)?$`)

// Start starts the debug shell for the given code runner.
func Start(codeRunner *coderunner.CodeRunner) {
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

		var commandValid bool = false

		switch command {
		case "r", "run":
			codeRunner.Run()
			fmt.Print("\n")
			commandValid = true

		case "c", "continue":
			codeRunner.Continue()
			fmt.Print("\n")
			commandValid = true

		case "l", "list":
			codeRunner.PrintBreakPoint()
			commandValid = true

		case "s", "see":
			codeRunner.PrintCode()
			commandValid = true

		case "h", "help":
			fmt.Print(HELP_STRING)
			commandValid = true

		case "q", "quit":
			return
		}

		if matches := REG_BREAK.FindStringSubmatch(command); matches != nil {
			// regex match break command
			commandValid = true

			var line uint64
			fmt.Sscanf(matches[2], "%d", &line)
			err := codeRunner.AddBreakPoint(line)

			if err != nil {
				fmt.Printf("%s\n\n", err.Error())
			}
		} else if matches := REG_DELETE.FindStringSubmatch(command); matches != nil {
			// regex match delete command
			commandValid = true

			var index int
			fmt.Sscanf(matches[2], "%d", &index)

			err := codeRunner.RemoveBreakPoint(index)
			if err != nil {
				fmt.Printf("%s\n\n", err.Error())
			}
		} else if matches := REG_PEEK.FindStringSubmatch(command); matches != nil {
			// regex match peek command
			commandValid = true

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

			var bytes []byte = codeRunner.PeekBytes(offset, length)
			for _, each := range bytes {
				fmt.Printf("%d ", each)
			}
			fmt.Print("\n")
		}

		// if no match command, print help
		if !commandValid {
			fmt.Print("\nUnknown command.\n")
			fmt.Print(HELP_STRING)
		}
	}
}
