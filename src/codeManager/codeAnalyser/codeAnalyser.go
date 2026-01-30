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
	"strings"

	"github.com/Anslen/Bfck/codeManager/bracketNotCloseError"
	"github.com/Anslen/Bfck/codeManager/code"
)

// Analyse analyses the given code text and returns a Code structure or an error.
func Analyse(codeText string, debugFlag bool) (ret *code.Code, err error) {
	// Create empty Code structure
	ret = code.New(debugFlag)

	// Return early if codeText is empty
	if len(codeText) == 0 {
		return
	}

	// Initialise variables
	var bracketIndexStack []uint64 = make([]uint64, 0)
	var lineCount uint64 = 0
	var lastOperator code.Operator = code.Invalid

	// Lookup each character in codeText
	for line := range strings.Lines(codeText) {
		lineCount++
		if debugFlag {
			ret.LineBegins = append(ret.LineBegins, len(ret.Operators))
		}
		var lineIsEmpty bool = true // Line is empty until an operator is found

		for columnIndex, char := range line {
			var forceNewOperator bool = (debugFlag && lineIsEmpty)
			switch char {
			case '+':
				// Try to combine with last operator
				// If current line is empty and in debug mode, force to create new operator
				if !forceNewOperator {
					if lastOperator == code.OpAdd {
						ret.Auxiliary[len(ret.Auxiliary)-1]++
						continue
					} else if lastOperator == code.OpSub {
						// If last operator is Sub, try to reduce it
						ret.Auxiliary[len(ret.Auxiliary)-1]--
						if ret.Auxiliary[len(ret.Auxiliary)-1] == 0 {
							// Remove last operator
							ret.Operators = ret.Operators[:len(ret.Operators)-1]
							ret.Auxiliary = ret.Auxiliary[:len(ret.Auxiliary)-1]

							// Reset lineIsEmpty if needed
							if debugFlag && ret.LineBegins[len(ret.LineBegins)-1] == len(ret.Operators) {
								lineIsEmpty = true
							}

							// Update lastOperator
							if len(ret.Operators) == 0 {
								lastOperator = code.Invalid
							} else {
								lastOperator = ret.Operators[len(ret.Operators)-1]
							}
						}
						continue
					}
				}
				pushOperator(ret, code.OpAdd)
				lastOperator = code.OpAdd
				lineIsEmpty = false

			case '-':
				if !forceNewOperator {
					if lastOperator == code.OpSub {
						ret.Auxiliary[len(ret.Auxiliary)-1]++
						continue

					} else if lastOperator == code.OpAdd {
						// If last operator is Add, try to reduce it
						ret.Auxiliary[len(ret.Auxiliary)-1]--
						if ret.Auxiliary[len(ret.Auxiliary)-1] == 0 {
							// Remove last operator
							ret.Operators = ret.Operators[:len(ret.Operators)-1]
							ret.Auxiliary = ret.Auxiliary[:len(ret.Auxiliary)-1]

							// Reset lineIsEmpty if needed
							if debugFlag && ret.LineBegins[len(ret.LineBegins)-1] == len(ret.Operators) {
								lineIsEmpty = true
							}

							// Update lastOperator
							if len(ret.Operators) == 0 {
								lastOperator = code.Invalid
							} else {
								lastOperator = ret.Operators[len(ret.Operators)-1]
							}
						}
						continue
					}
				}
				pushOperator(ret, code.OpSub)
				lastOperator = code.OpSub
				lineIsEmpty = false

			case '<':
				if (lastOperator != code.OpMoveLeft) || forceNewOperator {
					pushOperator(ret, code.OpMoveLeft)
				} else {
					ret.Auxiliary[len(ret.Auxiliary)-1]++
				}
				lastOperator = code.OpMoveLeft
				lineIsEmpty = false

			case '>':
				if (lastOperator != code.OpMoveRight) || forceNewOperator {
					pushOperator(ret, code.OpMoveRight)
				} else {
					ret.Auxiliary[len(ret.Auxiliary)-1]++
				}
				lastOperator = code.OpMoveRight
				lineIsEmpty = false

			case '[':
				// Push breacket index onto stack
				bracketIndexStack = append(bracketIndexStack, uint64(len(ret.Operators)))
				pushOperator(ret, code.OpLeftBracket) // Auxiliary will be set later
				lastOperator = code.OpLeftBracket
				lineIsEmpty = false

			case ']':
				// Pop bracket index from stack
				if len(bracketIndexStack) == 0 {
					ret = nil
					err = bracketNotCloseError.New(lineCount, columnIndex, line)
					return
				}
				var leftBracketIndex uint64 = bracketIndexStack[len(bracketIndexStack)-1]
				bracketIndexStack = bracketIndexStack[:len(bracketIndexStack)-1]

				// Set jump positions in Auxiliary data
				pushOperator(ret, code.OpRightBracket)
				ret.Auxiliary[len(ret.Auxiliary)-1] = leftBracketIndex

				// Set jump position for left bracket
				ret.Auxiliary[leftBracketIndex] = uint64(len(ret.Operators) - 1)

				lastOperator = code.OpRightBracket
				lineIsEmpty = false

			case ',':
				pushOperator(ret, code.OpInput)
				lastOperator = code.OpInput
				lineIsEmpty = false

			case '.':
				pushOperator(ret, code.OpOutput)
				lastOperator = code.OpOutput
				lineIsEmpty = false
			}
		}
	}
	ret.LineCount = lineCount
	ret.CodeCount = len(ret.Operators)

	// Set unvalid line begins to -1
	if debugFlag {
		for i := len(ret.LineBegins) - 1; i >= 0; i-- {
			if ret.LineBegins[i] >= len(ret.Operators) {
				ret.LineBegins[i] = -1
			} else {
				break
			}
		}
	}

	// Check for unclosed brackets
	if len(bracketIndexStack) != 0 {
		ret = nil
		// Get position of unclosed bracket
		var unclosedBracketCount int = len(bracketIndexStack)
		lineCount = 0
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
	}

	return
}

// pushOperator appends an operator, its Auxiliary will be set to 1.
func pushOperator(code *code.Code, op code.Operator) {
	code.Operators = append(code.Operators, op)
	code.Auxiliary = append(code.Auxiliary, 1)
}
