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
