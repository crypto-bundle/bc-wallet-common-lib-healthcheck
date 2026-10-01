// MIT NON-AI License
//
// Copyright (c) 2022-2026 Aleksei Kotelnikov(gudron2s@gmail.com)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of the software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions.
//
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
// In addition, the following restrictions apply:
//
// 1. The Software and any modifications made to it may not be used for the purpose of training or improving machine learning algorithms,
// including but not limited to artificial intelligence, natural language processing, or data mining. This condition applies to any derivatives,
// modifications, or updates based on the Software code. Any usage of the Software in an AI-training dataset is considered a breach of this License.
//
// 2. The Software may not be included in any dataset used for training or improving machine learning algorithms,
// including but not limited to artificial intelligence, natural language processing, or data mining.
//
// 3. Any person or organization found to be in violation of these restrictions will be subject to legal action and may be held liable
// for any damages resulting from such use.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
// OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package healthcheck

import (
	"sync"
)

const defaultStatesListSize = 8

var _ probeStatesListService = (*probeStatesList)(nil)

// probeStatesList stores up to N historical ProbeUnitResult snapshots per named state.
// It appends new states and evicts the oldest when the buffer exceeds its capacity.
type probeStatesList struct {
	mu *sync.RWMutex

	states []*ProbeUnitResult
	size   int
}

// NewProbeStateList creates a FIFO list with the given buffer size.
func NewProbeStateList(bufferSize int) *probeStatesList {
	return &probeStatesList{
		mu:     &sync.RWMutex{},
		states: make([]*ProbeUnitResult, 0, bufferSize),
		size:   bufferSize,
	}
}

// AddState deep-clones the input and appends it to the end of the list,
// evicting the oldest entry if the buffer exceeds its configured size.
func (w *probeStatesList) AddState(state *ProbeUnitResult) {
	if state == nil {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	cloned := state.Clone()
	w.states = append(w.states, &cloned)

	if len(w.states) > w.size {
		w.states = w.states[len(w.states)-w.size:]
	}
}

// GetLastState returns a shallow copy of the most recent snapshot pointer,
// or nil if the list is empty. The underlying ProbeUnitResult is already
// deep-cloned at insertion time by AddState.
func (w *probeStatesList) GetLastState() *ProbeUnitResult {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getLastStateLocked()
}

func (w *probeStatesList) getLastStateLocked() *ProbeUnitResult {
	if len(w.states) == 0 {
		return nil
	}

	return w.states[len(w.states)-1]
}

// GetStatesCount returns the number of stored snapshots.
func (w *probeStatesList) GetStatesCount() uint {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return uint(len(w.states))
}

// GetAllStates returns a newly allocated slice containing deep clones
// of every stored snapshot. Callers may safely mutate the returned slice
// or its contents without affecting internal state.
func (w *probeStatesList) GetAllStates() []*ProbeUnitResult {
	w.mu.RLock()
	defer w.mu.RUnlock()

	res := make([]*ProbeUnitResult, len(w.states))
	for i, s := range w.states {
		c := *s
		res[i] = &c
	}

	return res
}

// DrainAllStates atomically captures a deep clone of all current snapshots,
// clears the internal list, and returns the captured copy. The internal list
// is empty after this call.
func (w *probeStatesList) DrainAllStates() []*ProbeUnitResult {
	w.mu.Lock()
	defer w.mu.Unlock()

	res := make([]*ProbeUnitResult, len(w.states))
	for i, s := range w.states {
		c := *s
		res[i] = &c
	}

	w.states = w.states[:0]

	return res
}
