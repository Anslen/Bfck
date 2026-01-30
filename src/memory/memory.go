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

package memory

const MemoryBlockSize = 1024

type Memory struct {
	cells []byte
	ptr   int
	prev  *Memory
	next  *Memory
}

func New() (ret *Memory) {
	ret = &Memory{
		cells: make([]byte, MemoryBlockSize),
		ptr:   MemoryBlockSize / 2,
	}
	return
}

// Peek returns the byte at the current pointer plus the given offset.
func (m *Memory) Peek(offset int) (ret byte) {
	var index int = m.ptr + offset
	var current *Memory = m

	for index < 0 {
		if current.prev == nil {
			return 0
		}
		index += MemoryBlockSize
		current = current.prev
	}

	for index >= MemoryBlockSize {
		if current.next == nil {
			return 0
		}
		index -= MemoryBlockSize
		current = current.next
	}

	return current.cells[index]
}

// PeekBytes returns a slice of bytes starting from the current pointer plus the given offset.
func (m *Memory) PeekBytes(offset, length int) (ret []byte) {
	ret = make([]byte, length)
	for i := 0; i < length; i++ {
		ret[i] = m.Peek(offset + i)
	}
	return
}

// Poke sets the byte at the current pointer to the given value.
func (m *Memory) Poke(value byte) {
	m.cells[m.ptr] = value
}

// Add adds the given value to the byte at the current pointer.
func (m *Memory) Add(value uint64) {
	m.cells[m.ptr] += byte(value)
}

// Sub subtracts the given value from the byte at the current pointer.
func (m *Memory) Sub(value uint64) {
	m.cells[m.ptr] -= byte(value)
}

// MovePtr moves the pointer by the given offset, returning the Memory block where the pointer ends up.
//
// WARNING: Old Memory maybe invalid after calling this function.
func (m *Memory) MovePtr(offset int) (ret *Memory) {
	ret = m
	ret.ptr += offset

	// Check bounds and move to next/prev block if necessary
	for ret.ptr < 0 {
		if ret.prev == nil {
			ret.prev = New()
			ret.prev.next = ret
		}
		ret.prev.ptr = ret.ptr + MemoryBlockSize
		ret = ret.prev
	}
	for ret.ptr >= MemoryBlockSize {
		if ret.next == nil {
			ret.next = New()
			ret.next.prev = ret
		}
		ret.next.ptr = ret.ptr - MemoryBlockSize
		ret = ret.next
	}

	return
}
