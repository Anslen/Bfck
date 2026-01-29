package codereader

import (
	"os"

	"github.com/Anslen/Bfck/codeManager/code"
	codeanalyser "github.com/Anslen/Bfck/codeManager/codeAnalyser"
	coderunner "github.com/Anslen/Bfck/codeManager/codeRunner"
)

// Read reads the code from the given file path and returns a Code object.
func Read(path string, debugFlag bool) (ret *coderunner.CodeRunner, err error) {
	// Read file
	codeBytes, err := os.ReadFile(path)
	if err != nil {
		return
	}
	codeText := string(codeBytes)

	// Analyse code
	var code *code.Code
	code, err = codeanalyser.Analyse(codeText, debugFlag)
	if err != nil {
		return
	}

	// Create code runner
	ret = coderunner.New(code, debugFlag)

	return
}
