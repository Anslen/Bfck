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
	if m.ptr+offset < 0 {
		if m.prev == nil {
			m.prev = New()
			m.prev.next = m
		}
		return m.prev.Peek(m.ptr + offset + MemoryBlockSize)
	}

	if m.ptr+offset >= MemoryBlockSize {
		if m.next == nil {
			m.next = New()
			m.next.prev = m
		}
		return m.next.Peek(m.ptr + offset - MemoryBlockSize)
	}

	return m.cells[m.ptr+offset]
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
