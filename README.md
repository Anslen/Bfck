# Bfck

**Bfck** is a high-performance Brainfuck interpreter written in Go, featuring a powerful built-in interactive debugger. It not only supports standard Brainfuck code execution but also provides a GDB-like debugging experience, including breakpoints, memory watching, stepping, and disassembly views.

**How it works**: The interpreter performs static analysis on the source code to generate auxiliary arrays. It optimizes execution by merging adjacent mergeable operations (such as multiple `+` or `>`) and pre-calculating jump destinations for bracket loops. This approach significantly reduces runtime overhead and improves efficiency.

## Features

*   **Interactive Debug Shell**: Built-in debugger to debug code without extra tools.
*   **Breakpoint Management**: Support for setting and deleting breakpoints at specific line numbers.
*   **Memory Watch**: Real-time monitoring of specific memory cell changes.
*   **Code Analysis**: Ability to parse and view assembly-level instructions with auxiliary info in debug mode.
*   **Execution Control**: Supports stepping (`step`), running until loop end (`until`), and continuing execution (`continue`).

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

## Debugging Commands

After entering debug mode (seeing the `(Bfck)` prompt), use the following commands to control program execution. Characters in brackets `[]` indicate optional shorthand inputs.

| Command    | Short | Arguments           | Description                                                                                      |
| :--------- | :---- | :------------------ | :----------------------------------------------------------------------------------------------- |
| `run`      | `r`   | None                | Run code from the beginning.                                                                     |
| `continue` | `c`   | None                | Continue execution until the next breakpoint or program end.                                     |
| `step`     | `s`   | None                | Execute the next instruction (single step).                                                      |
| `until`    | `u`   | None                | Run until the current loop `[]` finishes.                                                        |
| `break`    | `b`   | `<line>`            | Set a breakpoint at the specified line number. E.g., `b 10`.                                     |
| `delete`   | `d`   | `<line>`            | Delete the breakpoint at the specified line number.                                              |
| `watch`    | `w`   | `<offset>`          | Watch the memory value at the current pointer relative offset. E.g., `w 0` watches current cell. |
| `peek`     | `p`   | `[offset [length]]` | Peek memory data. Defaults to current cell. E.g., `p 0 5` peeks 5 bytes starting from current.   |
| `info`     | `i`   | None                | Show current breakpoints and watch list.                                                         |
| `next`     | `n`   | None                | Show the next operator to be executed.                                                           |
| `code`     | None  | None                | Show the full list of parsed code instructions.                                                  |
| `clear`    | None  | None                | Clear all breakpoints.                                                                           |
| `help`     | `h`   | None                | Show help message.                                                                               |
| `quit`     | `q`   | None                | Quit the debugger.                                                                               |

### Auxiliary Data

When using the `code` command or viewing instructions, you may see an **Auxiliary** value associated with each operator. This is the result of the interpreter's optimization:

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
