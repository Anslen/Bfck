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

package codeanalyser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Anslen/Bfck/codeManager/bracketNotCloseError"
	"github.com/Anslen/Bfck/codeManager/code"
)

type analyser struct {
	debugFlag         bool
	lineCount         int
	columnIndex       int
	currentLine       string
	lineIsEmpty       bool
	lastOperator      code.Operator
	bracketIndexStack []uint64
}

// Analyse analyses the given code text and returns a Code structure or an error.
func Analyse(codeText string, debugFlag bool) (ret *code.Code, err error) {
	// Create empty Code structure
	ret = code.New(debugFlag)

	// Return early if codeText is empty
	if len(codeText) == 0 {
		err = errors.New("Error: Code is empty")
		return
	}

	// Initialise variables
	var analyser *analyser = &analyser{
		debugFlag:         debugFlag,
		lineIsEmpty:       true,
		lastOperator:      code.Invalid,
		bracketIndexStack: make([]uint64, 0),
	}

	// Lookup each character in codeText
	for line := range strings.Lines(codeText) {
		// Record line begin
		if debugFlag {
			ret.LineBegins = append(ret.LineBegins, len(ret.Operators))
		}

		// New line
		analyser.currentLine = line
		analyser.lineCount++
		analyser.lineIsEmpty = true // Line is empty until an operator is found
		analyser.columnIndex = -1

		// Analyse each character in the line
		for _, char := range line {
			analyser.columnIndex++
			err = analyser.analyseChar(ret, char)
			if err != nil {
				ret = nil
				return
			}
		}
	}

	// Set final counts
	ret.LineCount = uint64(analyser.lineCount)
	ret.CodeCount = len(ret.Operators)

	// Adjust line begins
	if debugFlag {
		adjustLineBegins(ret)
	}

	// Check for unclosed brackets
	err = analyser.checkBracketMatch(codeText)
	if err != nil {
		ret = nil
		return
	}

	if len(ret.Operators) == 0 {
		// Code is empty after analysis
		ret = nil
		err = errors.New("Error: Code is empty")
	}

	return
}

// analyseChar analyses a single character and updates the Code structure accordingly.
func (a *analyser) analyseChar(result *code.Code, char rune) (err error) {
	op := code.ToOperator(char)
	switch op {
	case code.OpAdd, code.OpSub, code.OpMoveLeft, code.OpMoveRight:
		a.processSimpleOperator(result, op)

	case code.OpInput, code.OpOutput:
		pushOperator(result, op)
		a.lastOperator = op
		a.lineIsEmpty = false

	case code.OpLeftBracket:
		// Push breacket index onto stack
		a.bracketIndexStack = append(a.bracketIndexStack, uint64(len(result.Operators)))
		pushOperator(result, code.OpLeftBracket) // Auxiliary will be set later
		a.lastOperator = code.OpLeftBracket
		a.lineIsEmpty = false

	case code.OpRightBracket:
		pushOperator(result, code.OpRightBracket)

		// Set jump indices
		err = a.setJumpIndex(result)
		if err != nil {
			result = nil
			return
		}

		a.lastOperator = code.OpRightBracket
		a.lineIsEmpty = false
	}
	return nil
}

// processSimpleOperator processes simple operators (+, -, <, >).
func (a *analyser) processSimpleOperator(result *code.Code, op code.Operator) {
	// If current line is empty and in debug mode, force to create new operator
	var forceNewOperator bool = (a.debugFlag && a.lineIsEmpty)
	if !forceNewOperator {
		// Combine with last operator if possible
		if a.lastOperator == op {
			result.Auxiliary[len(result.Auxiliary)-1]++
			return

		} else if a.lastOperator == op.Reverse() {
			// If last operator is Reverse of op, reduce it
			a.reduceLastOperator(result)
			return
		}
	}
	pushOperator(result, op)
	a.lastOperator = op
	a.lineIsEmpty = false
}

// pushOperator appends an operator, its Auxiliary will be set to 1.
func pushOperator(result *code.Code, op code.Operator) {
	result.Operators = append(result.Operators, op)
	result.Auxiliary = append(result.Auxiliary, 1)
}

// reduceLastOperator reduces the last operator by 1, and removes it if Auxiliary becomes 0.
//
// Used to optimize consecutive opposite operators.
func (a *analyser) reduceLastOperator(result *code.Code) {
	// Reduce last operator by 1
	result.Auxiliary[len(result.Auxiliary)-1]--

	// If Auxiliary becomes 0, remove the operator
	if result.Auxiliary[len(result.Auxiliary)-1] == 0 {
		// Remove last operator
		result.Operators = result.Operators[:len(result.Operators)-1]
		result.Auxiliary = result.Auxiliary[:len(result.Auxiliary)-1]

		// Reset lineIsEmpty if needed
		if a.debugFlag && result.LineBegins[len(result.LineBegins)-1] == len(result.Operators) {
			a.lineIsEmpty = true
		}

		// Update lastOperator
		if len(result.Operators) == 0 {
			a.lastOperator = code.Invalid
		} else {
			a.lastOperator = result.Operators[len(result.Operators)-1]
		}
	}
}

// setJumpIndex sets the jump index for the brackets in the bracketIndexStack.
//
// Right bracket should be added to code before calling this function.
//
// line: current text line
func (a *analyser) setJumpIndex(result *code.Code) (err error) {
	// Pop bracket index from stack
	if len(a.bracketIndexStack) == 0 {
		err = bracketNotCloseError.New(a.lineCount, a.columnIndex, a.currentLine)
		return
	}
	var leftBracketIndex uint64 = a.bracketIndexStack[len(a.bracketIndexStack)-1]
	a.bracketIndexStack = a.bracketIndexStack[:len(a.bracketIndexStack)-1]

	// Check empty loop and warn
	if leftBracketIndex == uint64(len(result.Operators))-2 {
		fmt.Printf("Warning: Empty loop at line %v\n", a.lineCount)
	}

	// Set jump indices in Auxiliary data
	result.Auxiliary[len(result.Auxiliary)-1] = leftBracketIndex + 1
	result.Auxiliary[leftBracketIndex] = uint64(len(result.Operators))
	return nil
}

// adjustLineBegins adjusts the LineBegins slice to ensure all positions are valid.
func adjustLineBegins(result *code.Code) {
	// Reverse iterate lineBegins to set invalid positions to -1
	for i := len(result.LineBegins) - 1; i >= 0; i-- {
		if result.LineBegins[i] >= len(result.Operators) {
			result.LineBegins[i] = -1
		} else {
			return
		}
	}
}

// checkBracketMatch checks if there are any unclosed brackets in the bracketIndexStack.
func (a *analyser) checkBracketMatch(codeText string) (err error) {
	// Check for unclosed brackets
	if len(a.bracketIndexStack) == 0 {
		return nil
	}

	// Get position of unclosed bracket
	var unclosedBracketCount int = len(a.bracketIndexStack)
	var lineCount int = 0
	// Find line and column of unclosed bracket
	for line := range strings.Lines(codeText) {
		lineCount++
		for columnIndex, char := range line {
			if char == '[' {
				unclosedBracketCount--
				if unclosedBracketCount == 0 {
					err = bracketNotCloseError.New(lineCount, columnIndex, line)
					return
				}
			}
		}
	}

	panic("codeAnalyser: Unclosed bracket position not found")
}
