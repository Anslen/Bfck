# Bfck

**Bfck** is a high-performance Brainfuck interpreter written in Go, featuring a powerful built-in interactive debugger. It not only supports standard Brainfuck code execution but also provides a GDB-like debugging experience, including breakpoints, memory watching, stepping, and disassembly views.

**How it works**: The interpreter performs static analysis on the source code to generate auxiliary arrays. It optimizes execution by merging adjacent mergeable operations (such as multiple `+` or `>`) and pre-calculating jump destinations for bracket loops. This approach significantly reduces runtime overhead and improves efficiency.

## Features

*   **Breakpoint Management**: Support for setting and managing breakpoints with gdb-like commands.
*   **Memory Watch**: Real-time monitoring of specific memory cell changes.
*   **Code Analysis**: Ability to parse and view assembly-level instructions with auxiliary info and loop labels in debug mode.
*   **Execution Control**: Supports stepping (`step`), running until loop end (`until`), continuing execution (`continue`), and stopping at specific instruction (`stop`).
*   **Detailed Execution Visualization**: The `detailed` command visualizes each execution step, showing the current instruction and surrounding memory tape state.

## Quick Start

### Build

**Windows:**
```powershell
./compile/windows.ps1
```

**Linux:**
```bash
chmod +x compile/linux.sh
./compile/linux.sh
```

The executable will be generated in the `bin` directory.

### Usage

Show help message:
```bash
./bfck help
```

Run Brainfuck file directly:
```bash
./bfck run <file_path>
```

Enter debug mode:
```bash
./bfck debug <file_path>
```

**Note**: In debug mode, memory state is preserved after execution finishes for convenience checking. It will be automatically reset when you start a new run. You can use `reset` command to manually reset memory. Debug configurations like `watch` list are persistent and will NOT be cleared by this automatic reset or the manual `reset` command but will be cleared after running finish.

**Note**: When using `step` command to execute multiple instructions, the execution will be interrupted by **watch** memory, but it will ignore **breakpoints** and **stop instruction**.

### Example

Consider `example.bf`:
```brainfuck
++++
>++
[-]
```

Analysed code:
```text
$ ./bin/bfck debug example.bf
(Bfck) code

Total operators count: 6

Code with auxiliary:
  0        Add             4
  1        MoveRight       1
  2        Add             2
L1:
  3        LeftBracket     6
  4        Sub             1
  5        RightBracket    4
L1 End

Lines count: 3

Line begins at:
Line    begin
  1     0
  2     1
  3     3
```

Debug with `detailed` command:
```text
(Bfck) d
0        Add             4

Memory pointer at: 0
0 0 0 0 0 0 0 0 0 0 [4] 0 0 0 0 0 0 0 0 0

1        MoveRight       1

Memory pointer at: 0
0 0 0 0 0 0 0 0 0 4 [0] 0 0 0 0 0 0 0 0 0

2        Add             2

Memory pointer at: 1
0 0 0 0 0 0 0 0 0 4 [2] 0 0 0 0 0 0 0 0 0

3        LeftBracket     6

Memory pointer at: 1
0 0 0 0 0 0 0 0 0 4 [2] 0 0 0 0 0 0 0 0 0

4        Sub             1

Memory pointer at: 1
0 0 0 0 0 0 0 0 0 4 [1] 0 0 0 0 0 0 0 0 0

5        RightBracket    4

Memory pointer at: 1
0 0 0 0 0 0 0 0 0 4 [1] 0 0 0 0 0 0 0 0 0

4        Sub             1

Memory pointer at: 1
0 0 0 0 0 0 0 0 0 4 [0] 0 0 0 0 0 0 0 0 0

5        RightBracket    4

Memory pointer at: 1
0 0 0 0 0 0 0 0 0 4 [0] 0 0 0 0 0 0 0 0 0



Running finished
```



## Debugging Commands

After entering debug mode (seeing the `(Bfck)` prompt), use the following commands to control program execution. Characters in brackets `[]` indicate optional shorthand inputs.

| Command    | Short | Arguments           | Description                                                                                      |
| :--------- | :---- | :------------------ | :----------------------------------------------------------------------------------------------- |
| `run`      | `r`   | None                | Run code from the beginning.                                                                     |
| `continue` | `c`   | None                | Continue execution until the next breakpoint or program end.                                     |
| `step`     | `s`   | `[times]`           | Execute the next instruction (single step), or multiple times if specified. Will be interrupted by watchpoints but ignores breakpoints and stop instructions. |
| `detailed` | `d`   | `[times]`           | Execute detailed steps (default 1), showing the instruction and memory tape after each step.     |
| `until`    | `u`   | None                | Run until the current loop `[]` finishes.                                                        |
| `stop`     | None  | `<index>`           | Stop execution at the specified operator index.                                                  |
| `tape`     | `t`   | None                | Show memory tape around current pointer.                                                         |
| `ptr`      | None  | None                | Show the current memory pointer address (Start is 0).                                            |
| `break`    | `b`   | `<line>`            | Set a breakpoint at the specified line number. E.g., `b 10`.                                     |
| `delete`   | `del` | `s\|b\|w <num>`     | Delete the stop point (`s`), breakpoint (`b`) or watchpoint (`w`) at the specified index.        |
| `watch`    | `w`   | `<address>`         | Watch the memory at the specified absolute address. E.g., `w 0` watches the starting cell.       |
| `peek`     | `p`   | `[offset [length]]` | Peek memory data. Defaults to current cell. E.g., `p 0 5` peeks 5 bytes starting from current.   |
| `info`     | `i`   | `[s\|b\|w]`         | Show current stop points (`s`), breakpoints (`b`) or watch list (`w`). Default shows all.        |
| `next`     | `n`   | None                | Show the next operator to be executed.                                                           |
| `reset`    | None  | None                | Manually reset memory and execution state.                                                       |
| `code`     | None  | None                | Show the full list of parsed code instructions.                                                  |
| `clear`    | None  | `[s\|b\|w]`         | Clear stop points (`s`), breakpoints (`b`) or watchpoints (`w`). Default clears all.             |
| `help`     | `h`   | None                | Show help message.                                                                               |
| `quit`     | `q`   | None                | Quit the debugger.                                                                               |

### Auxiliary Data

When using the `code` command or viewing instructions, you will see an **Auxiliary** value associated with each operator. This is the result of the interpreter's optimization:

*   **`+` / `-` (Add / Sub)**: The number of times to increment/decrement (e.g., `+++` becomes `Add 3`).
*   **`>` / `<` (MoveRight / MoveLeft)**: The number of steps to move the pointer.
*   **`[` / `]` (LeftBracket / RightBracket)**: The index of the matching bracket to jump to.
*   **`.` / `,` (Output / Input)**: Usually 1.

## Brainfuck Language Reference

This project supports standard Brainfuck syntax:

| Char | Meaning                                                                                   |
| :--: | :---------------------------------------------------------------------------------------- |
| `>`  | Move the pointer to the right.                                                            |
| `<`  | Move the pointer to the left.                                                             |
| `+`  | Increment the byte at the pointer.                                                        |
| `-`  | Decrement the byte at the pointer.                                                        |
| `.`  | Output the byte at the pointer as an ASCII character.                                     |
| `,`  | Accept one byte of input, storing its value in the byte at the pointer.                   |
| `[`  | If the byte at the pointer is zero, jump it forward to the command after the matching `]`.|
| `]`  | If the byte at the pointer is nonzero, jump it back to the command after the matching `[`.|

## Contributing

Contributions via Issues or Pull Requests are welcome!

1.  **Memory Model**:
    The memory tape is implemented as an infinitely extendable **doubly-linked list**. Each node (block) has a capacity of **1024 bytes**. In most cases, a single block is sufficient, so extra allocation and linked list traversal overheads are rarely triggered.

2.  **Indexing Convention**:
    *   **0-based**: Internal arrays (like `Operators`, `Auxiliary` in `Code` struct) and memory offsets use 0-based indexing.
    *   **1-based**: User-facing line numbers (e.g., in debug commands `break <line>`) are 1-based to align with text editors.
    Please be mindful of this distinction when modifying the debugger or code manager.

## License

[GNU General Public License](LICENSE)
