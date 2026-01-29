package main

import (
	"fmt"
	"os"

	codereader "github.com/Anslen/Bfck/codeManager/codeReader"
	debugshell "github.com/Anslen/Bfck/debugShell"
)

const MAIN_DEBUG = true
const MAIN_DEBUG_FILE_PATH = "C:/Codes/go/Bfck/bfSrc/print.bf"

const HELP_STRING string = "run <file_path>   : Run specified code file without debug\n" +
	"debug <file_path> : Open debug shell with specified code file\n" +
	"help              : Show this help message\n"

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
