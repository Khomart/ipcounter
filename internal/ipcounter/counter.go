package ipcounter

import "sync/atomic"

type IPCounter struct {
	existingIPs []byte
	uniqueIPs   atomic.Int32
}

func New() *IPCounter {
	return &IPCounter{
		existingIPs: make([]byte, 1<<32/8),
	}
}

func (h *IPCounter) Add(ip uint32) {
	bIndex := ip / 8
	oIndex := ip % 8
	if h.existingIPs[bIndex]&(1<<oIndex) == 0 {
		h.existingIPs[bIndex] |= 1 << oIndex
		h.uniqueIPs.Add(1)
	}
}

func (h *IPCounter) GetUniqueCount() int32 {
	return h.uniqueIPs.Load()
}
