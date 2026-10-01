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
	"testing"
)

func TestNewProbeStateList_EmptyDefault(t *testing.T) {
	t.Parallel()

	list := NewProbeStateList(defaultStatesListSize)

	if list == nil {
		t.Fatal("expected non-nil list")
	}

	if list.GetStatesCount() != 0 {
		t.Errorf("expected count 0, got %d", list.GetStatesCount())
	}

	if list.GetLastState() != nil {
		t.Error("expected nil last state for empty list")
	}
}

func TestNewProbeStateList_CustomBufferSize(t *testing.T) {
	t.Parallel()

	const customSize = 3

	list := NewProbeStateList(customSize)

	if list.size != customSize {
		t.Errorf("expected buffer size %d, got %d", customSize, list.size)
	}
}

func TestProbeStateList_AddState_ClonedInternally(t *testing.T) {
	t.Parallel()

	list := NewProbeStateList(8)
	original := &ProbeUnitResult{Name: "test", Message: "hello", Code: 42}

	list.AddState(original)

	last := list.GetLastState()
	if last == nil {
		t.Fatal("expected non-nil last state after add")
	}

	// Mutate original — stored value must be unaffected.
	original.Message = "mutated"

	if last.Message != "hello" {
		t.Errorf("expected cloned message %q, got %q", "hello", last.Message)
	}
}

func TestProbeStateList_AddState_NilSkipped(t *testing.T) {
	t.Parallel()

	list := NewProbeStateList(8)
	list.AddState(nil)

	if count := list.GetStatesCount(); count != 0 {
		t.Errorf("expected count 0 after nil add, got %d", count)
	}

	if last := list.GetLastState(); last != nil {
		t.Error("expected nil last state after nil add")
	}
}

func TestProbeStateList_AddState_AppendOrder(t *testing.T) {
	t.Parallel()

	list := NewProbeStateList(8)
	names := []string{"first", "second", "third"}

	for _, name := range names {
		list.AddState(&ProbeUnitResult{Name: name})
	}

	// Append order: last should be "third" (newest at end).
	last := list.GetLastState()
	if last == nil || last.Name != "third" {
		t.Errorf("expected newest state name 'third', got %+v", last)
	}

	all := list.GetAllStates()
	for i, s := range all {
		if s.Name != names[i] {
			t.Errorf("at index %d expected name %q, got %q", i, names[i], s.Name)
		}
	}
}

func TestProbeStateList_GetAllStates_ReturnsCopies(t *testing.T) {
	t.Parallel()

	list := NewProbeStateList(8)
	list.AddState(&ProbeUnitResult{Name: "x", Message: "val"})

	all := list.GetAllStates()
	if len(all) != 1 {
		t.Fatalf("expected 1 state, got %d", len(all))
	}

	// Mutate returned slice contents.
	all[0].Message = "mutated"
	all[0].Name = "changed"

	// Second fetch must be unaffected.
	all2 := list.GetAllStates()
	if all2[0].Message != "val" || all2[0].Name != "x" {
		t.Error("returned slice was not a deep clone")
	}
}

func TestProbeStateList_GetLastState_EmptyReturnsNil(t *testing.T) {
	t.Parallel()

	list := NewProbeStateList(8)

	if last := list.GetLastState(); last != nil {
		t.Error("expected nil on empty list")
	}
}

func TestProbeStateList_DrainAllStates_ClearsInternal(t *testing.T) {
	t.Parallel()

	list := NewProbeStateList(8)
	list.AddState(&ProbeUnitResult{Name: "a"})
	list.AddState(&ProbeUnitResult{Name: "b"})

	captured := list.DrainAllStates()

	if count := list.GetStatesCount(); count != 0 {
		t.Errorf("expected drain to clear internal list, got count %d", count)
	}

	if captured == nil || len(captured) != 2 {
		t.Fatalf("expected 2 drained states, got %v", captured)
	}

	if captured[0].Name != "a" || captured[1].Name != "b" {
		t.Error("drained states have wrong names")
	}
}

func TestProbeStateList_DrainAllStates_ReturnsSnapshot(t *testing.T) {
	t.Parallel()

	list := NewProbeStateList(8)
	list.AddState(&ProbeUnitResult{Name: "drain-me", Message: "value"})

	captured := list.DrainAllStates()

	// Drain is idempotent — second drain returns nothing.
	if after := list.DrainAllStates(); len(after) != 0 {
		t.Errorf("expected empty after first drain, got %d states", len(after))
	}

	// Captured data must survive subsequent drains.
	if captured[0].Name != "drain-me" || captured[0].Message != "value" {
		t.Error("captured snapshot was corrupted")
	}
}

func TestProbeStateList_EvictionOnOverflow(t *testing.T) {
	t.Parallel()

	const bufSize = 3

	list := NewProbeStateList(bufSize)

	// Insert more items than buffer can hold.
	names := []string{"a", "b", "c", "d", "e"}
	for _, n := range names {
		list.AddState(&ProbeUnitResult{Name: n})
	}

	count := list.GetStatesCount()
	if count != 3 {
		t.Errorf("expected buffer count 3 after overflow, got %d", count)
	}

	// Oldest entries (a, b) must be evicted; remaining: c, d, e.
	all := list.GetAllStates()
	expectedNames := []string{"c", "d", "e"}

	for i, s := range all {
		if s.Name != expectedNames[i] {
			t.Errorf("at index %d expected name %q, got %q", i, expectedNames[i], s.Name)
		}
	}

	// Newest should be "e".
	last := list.GetLastState()
	if last == nil || last.Name != "e" {
		t.Errorf("expected newest name 'e', got %+v", last)
	}
}

func TestProbeStateList_ConcurrentReadWriteRaceDetector(t *testing.T) {
	t.Parallel()

	list := NewProbeStateList(16)

	var wg sync.WaitGroup

	// 5 writers.
	for id := range 5 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for range 100 {
				list.AddState(&ProbeUnitResult{
					Name:    string(rune('a' + id)),
					Message: "",
					Code:    uint16(id),
				})
			}
		}(id)
	}

	// 5 readers.
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for range 100 {
				_ = list.GetLastState()
				_ = list.GetAllStates()
				_ = list.GetStatesCount()
			}
		}()
	}

	// 2 drainers.
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for range 50 {
				_ = list.DrainAllStates()
			}
		}()
	}

	wg.Wait()
}
