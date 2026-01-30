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

package main

import (
	"fmt"
	"os"

	codereader "github.com/Anslen/Bfck/codeManager/codeReader"
	debugshell "github.com/Anslen/Bfck/debugShell"
)

// For IDE debug
const MAIN_DEBUG = false
const MAIN_DEBUG_FILE_PATH = ""

const HELP_STRING string = "run <file_path>   : Run specified code file without debug\n" +
	"debug <file_path> : Open debug shell with specified code file\n" +
	"help              : Show this help message\n"

const VERSION_STRING string = "Bfck version 0.0.1 - Copyright (C) 2026 Anslen"

func main() {
	if MAIN_DEBUG {
		codeRunner, err := codereader.Read(MAIN_DEBUG_FILE_PATH, true)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		debugshell.Start(codeRunner)
		return
	}

	if len(os.Args) == 2 && os.Args[1] == "help" {
		fmt.Print(HELP_STRING)
		return
	}

	if len(os.Args) != 3 {
		fmt.Println("Unknown command. type 'help' for help.")
		return
	}

	switch os.Args[1] {
	case "run":
		codeRunner, err := codereader.Read(os.Args[2], false)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		codeRunner.Run()
	case "debug":
		codeRunner, err := codereader.Read(os.Args[2], true)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		debugshell.Start(codeRunner)

	default:
		fmt.Println("Unknown command. type 'help' for help.")
	}
}
