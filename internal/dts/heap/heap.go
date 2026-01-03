package heap

type Cmp[T any] func(a, b T) int

type Heap[T any] struct {
	cmp  Cmp[T]
	base []T n    int
}

func (h *Heap[T]) Size() int {
	return h.n
}

func (h *Heap[T]) Push(v T) {
	h.base = append(h.base, v)
	idx := h.Size() - 1
	for idx > 0 {
		pIdx := (idx - 1) / 2
		if h.cmp(h.base[pIdx], h.base[idx]) <= 0 {
			break
		}
		h.base[pIdx], h.base[idx] = h.base[idx], h.base[pIdx]
		idx = pIdx
	}
	h.n++
}

func (h *Heap[T]) Pop() (T, bool) {
	var zero T
	if h.Size() == 0 {
		return zero, false
	}

	h.base[0], h.base[h.Size()-1] = h.base[h.Size()-1], h.base[0]
	v := h.base[h.Size()-1]
	h.base = h.base[:h.Size()-1]
	h.n--
	idx := 0
	for idx < h.Size() {
		lIdx := 2*idx + 1
		rIdx := 2*idx + 2
		smallestIdx := idx

		if lIdx < h.Size() && h.cmp(h.base[lIdx], h.base[smallestIdx]) < 0 {
			smallestIdx = lIdx
		}
		if rIdx < h.Size() && h.cmp(h.base[rIdx], h.base[smallestIdx]) < 0 {
			smallestIdx = rIdx
		}
		if smallestIdx == idx {
			break
		}
		h.base[smallestIdx], h.base[idx] = h.base[idx], h.base[smallestIdx]
		idx = smallestIdx
	}
	return v, true
}

func (h *Heap[T]) Peek() (T, bool) {
	if h.Size() == 0 {
		var zero T
		return zero, false
	}
	return h.base[0], true
}

func New[T any](cmp Cmp[T]) *Heap[T] {
	return &Heap[T]{
		cmp:  cmp,
		base: make([]T, 0),
		n:    0,
	}
}

func FromSlice[T any](cmp Cmp[T], data []T) *Heap[T] {
	// Heapify
	h := &Heap[T]{
		cmp:  cmp,
		base: make([]T, len(data)),
		n:    len(data),
	}
	copy(h.base, data)

	for i := (h.Size() - 2) / 2; i >= 0; i-- {
		idx := i
		for idx < h.Size() {
			lIdx := 2*idx + 1
			rIdx := 2*idx + 2
			smallestIdx := idx

			if lIdx < h.Size() && h.cmp(h.base[lIdx], h.base[smallestIdx]) < 0 {
				smallestIdx = lIdx
			}
			if rIdx < h.Size() && h.cmp(h.base[rIdx], h.base[smallestIdx]) < 0 {
				smallestIdx = rIdx
			}
			if smallestIdx == idx {
				break
			}
			h.base[smallestIdx], h.base[idx] = h.base[idx], h.base[smallestIdx]
			idx = smallestIdx
		}
	}

	return h
}

func From[T any](cmp Cmp[T], vs ...T) *Heap[T] {
	return FromSlice(cmp, vs)
}
