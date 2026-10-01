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
	"errors"
	"sync"
)

var (
	ErrProbeNotReady  = errors.New("probe unit is not ready")
	ErrProbeUnhealthy = errors.New("probe unit is unhealthy")
)

var _ probeStatesService = (*probeStatesUnit)(nil)

// probeStatesUnit manages named health-state entries where each entry stores a FIFO history of snapshots
// via probeStatesList. It provides per-field error tracking, status recalculation, and time-travel queries.
type probeStatesUnit struct {
	e             errorFormatterService
	mu            *sync.RWMutex
	states        map[string]*probeStatesList
	mustBeSuccess []string
	status        probeUnitStatus
}

// NewProbeStateUnit creates a new probeStatesUnit pre-populated with required state fields.
// Each field is initialized with a single successful snapshot in its history buffer.
func NewProbeStateUnit(
	errFmtSvc errorFormatterService,
	requiredStateFields ...string,
) *probeStatesUnit {
	statesMap := make(map[string]*probeStatesList, len(requiredStateFields))

	for _, name := range requiredStateFields {
		list := NewProbeStateList(defaultStatesListSize)
		list.AddState(&ProbeUnitResult{
			Name:       name,
			Message:    "",
			Code:       uint16(MessageCodeEverythingOK),
			err:        nil,
			isRequired: true,
		})

		statesMap[name] = list
	}

	return &probeStatesUnit{
		mu: &sync.RWMutex{},
		e:  errFmtSvc,

		status:        ProbeUnitStatusSuccess,
		mustBeSuccess: requiredStateFields,

		states: statesMap,
	}
}

// GetStatus returns a snapshot of the probe unit status using the latest state from each
// named field's history list, along with all tracked latest states. Status is computed
// locally rather than cached to avoid data races under concurrent readers.
func (w *probeStatesUnit) GetStatus() *ProbeUnitStatusReport {
	w.mu.RLock()
	defer w.mu.RUnlock()

	states := w.getLatestStates()
	localStatus := w.computeStatusLocal(states)

	return &ProbeUnitStatusReport{
		Status: localStatus,
		States: states,
	}
}

// computeStatusLocal derives the overall status from a map of latest states without mutating internal state.
func (w *probeStatesUnit) computeStatusLocal(states map[string]ProbeUnitResult) probeUnitStatus {
	for _, r := range states {
		if r.isRequired && r.err != nil {
			return ProbeUnitStatusFailed
		}
	}

	return ProbeUnitStatusSuccess
}

// GetState retrieves the latest snapshot for a named state entry, returning nil if the key does not exist.
func (w *probeStatesUnit) GetState(name string) *ProbeUnitResult {
	w.mu.RLock()
	defer w.mu.RUnlock()

	value, ok := w.states[name]
	if !ok {
		return nil
	}

	return value.GetLastState()
}

// GetStatesByNames returns a map of the latest values for only the requested names, skipping absent keys.
func (w *probeStatesUnit) GetStatesByNames(names ...string) map[string]ProbeUnitResult {
	w.mu.RLock()
	defer w.mu.RUnlock()

	res := make(map[string]ProbeUnitResult, len(names))

	for _, v := range names {
		value, isExists := w.states[v]
		if !isExists {
			continue
		}

		last := value.GetLastState()
		if last == nil {
			continue
		}

		res[v] = last.Clone()
	}

	return res
}

// GetStates returns a copy of the latest state for every registered field.
func (w *probeStatesUnit) GetStates() map[string]ProbeUnitResult {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getLatestStates()
}

func (w *probeStatesUnit) getLatestStates() map[string]ProbeUnitResult {
	res := make(map[string]ProbeUnitResult, len(w.states))

	for k, v := range w.states {
		last := v.GetLastState()
		if last != nil {
			res[k] = last.Clone()
		}
	}

	return res
}

// SetStateMessage updates the Message field of an existing entry or creates a new optional one,
// appending a new snapshot to the name's history list.
func (w *probeStatesUnit) SetStateMessage(name string, message string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.setSnapshotLocked(name, message, 0, nil, false)
}

// SetStateError applies the given error to a named entry, creating a new snapshot appended
// to that name's history list. If err is nil and the name already exists, the error is cleared.
// If err is nil and the name does not exist, this is a no-op (same as the original behavior).
func (w *probeStatesUnit) SetStateError(name string, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	switch {
	case err != nil:
		// Error present: create/update the entry and append a new snapshot to history.
		// New entries are always required; existing entries inherit from mustBeSuccess membership.
		isRequired := false
		if _, ok := w.states[name]; !ok {
			isRequired = true
		} else {
			for _, n := range w.mustBeSuccess {
				if n == name {
					isRequired = true

					break
				}
			}
		}

		msg := err.Error()
		rawCode := w.e.ErrGetCode(err)
		code := uint16(MessageCodeEverythingOK) //nolint:gosec // rawCode is guaranteed >= 0 here
		if rawCode > 0 {
			code = uint16(rawCode)
		}

		w.setSnapshotLocked(name, msg, code, err, isRequired)

	case containsName(name, w.mustBeSuccess):
		// Clear error on an existing required field: find it and clear via SetStateMessage (which preserves it if it exists).
		if _, ok := w.states[name]; ok {
			w.setSnapshotLocked(name, "", 0, nil, false)
		}

	default:
		// Optional unknown name with nil error — no-op, matching original semantics.
	}
}

// setSnapshotLocked creates a new ProbeUnitResult snapshot and appends it to the named field's history list.
// It recalculates overall status after mutation.
func (w *probeStatesUnit) setSnapshotLocked(name string, message string, code uint16, err error, isRequired bool) {
	snapshot := &ProbeUnitResult{
		Name:       name,
		Message:    message,
		Code:       code,
		err:        err,
		isRequired: isRequired,
	}

	stateList, ok := w.states[name]
	if !ok {
		stateList = NewProbeStateList(defaultStatesListSize)
		w.states[name] = stateList
	}

	stateList.AddState(snapshot)
	w.recalculateStatusLocked()
}

// GetHistory returns all historical snapshots for a named state field.
func (w *probeStatesUnit) GetHistory(name string) []*ProbeUnitResult {
	w.mu.RLock()
	defer w.mu.RUnlock()

	value, ok := w.states[name]
	if !ok {
		return nil
	}

	return value.GetAllStates()
}

// GetHistoryCount returns the number of stored snapshots for a named state field.
func (w *probeStatesUnit) GetHistoryCount(name string) uint {
	w.mu.RLock()
	defer w.mu.RUnlock()

	value, ok := w.states[name]
	if !ok {
		return 0
	}

	return value.GetStatesCount()
}

// recalculateStatusLocked iterates over all stored probes and sets the unit status to Failed if any
// required-field's latest snapshot still carries an error; otherwise it resets to Success.
// Must be called under write lock (mu.Lock).
func (w *probeStatesUnit) recalculateStatusLocked() {
	w.status = ProbeUnitStatusSuccess

	for _, r := range w.states {
		last := r.getLastStateLocked()
		if last == nil {
			continue
		}

		if last.isRequired && last.err != nil {
			w.status = ProbeUnitStatusFailed

			return
		}
	}
}

// containsName checks whether the given name appears in a string slice.
func containsName(name string, items []string) bool {
	for _, v := range items {
		if v == name {
			return true
		}
	}

	return false
}
