package bracketNotCloseError

import "fmt"

type BracketNotCloseError struct {
	line        uint64
	errorString string
}

// Error implements the error interface for BracketNotCloseError.
func (e *BracketNotCloseError) Error() string {
	return fmt.Sprintf("Error: Bracket not close at line %v\n%v", e.line, e.errorString)
}

// newBracketNotCloseError creates a new BracketNotCloseError with the given line, column, and error line.
//
// CAUSION: lineCount start from 1, columnIndex start from 0
func New(lineCount uint64, columnIndex int, errorLine string) (ret *BracketNotCloseError) {
	ret = &BracketNotCloseError{
		line:        lineCount,
		errorString: errorLine,
	}
	// add arrow to indicate the column
	if errorLine[len(errorLine)-1] != '\n' {
		ret.errorString += "\n"
	}
	for i := 0; i < columnIndex; i++ {
		ret.errorString += " "
	}
	ret.errorString += "^\n"
	return
}
