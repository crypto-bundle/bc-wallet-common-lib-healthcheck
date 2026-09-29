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

// probeStatesUnit manages named health-state entries with per-field error tracking and status recalculation.
type probeStatesUnit struct {
	e             errorFormatterService
	mu            *sync.RWMutex
	states        map[string]*ProbeUnitResult
	mustBeSuccess []string
	status        probeUnitStatus
}

// NewProbeStateUnit creates a new probeStatesUnit pre-populated with required state fields. Each field is initialized to success.
func NewProbeStateUnit(
	errFmtSvc errorFormatterService,
	requiredStateFields ...string,
) *probeStatesUnit {
	statesMap := make(map[string]*ProbeUnitResult, len(requiredStateFields))

	for _, v := range requiredStateFields {
		statesMap[v] = &ProbeUnitResult{
			Name:    v,
			Message: "",
			Code:    uint16(MessageCodeEverythingOK),

			err:        nil,
			isRequired: true,
		}
	}

	return &probeStatesUnit{
		mu: &sync.RWMutex{},
		e:  errFmtSvc,

		status:        ProbeUnitStatusSuccess,
		mustBeSuccess: requiredStateFields,

		states: statesMap,
	}
}

// GetStatus returns a snapshot of the probe unit status along with all tracked states.
func (w *probeStatesUnit) GetStatus() *ProbeUnitStatusReport {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return &ProbeUnitStatusReport{
		Status: w.status,
		States: w.getStates(),
	}
}

// GetState retrieves a single named state entry, returning nil if the key does not exist.
func (w *probeStatesUnit) GetState(name string) *ProbeUnitResult {
	w.mu.RLock()
	defer w.mu.RUnlock()

	value, ok := w.states[name]
	if !ok {
		return nil
	}

	return value
}

// GetStatesByNames returns a map of copies for only the requested names, skipping absent keys.
func (w *probeStatesUnit) GetStatesByNames(names ...string) map[string]ProbeUnitResult {
	w.mu.RLock()
	defer w.mu.RUnlock()

	res := make(map[string]ProbeUnitResult, len(names))

	for _, v := range names {
		value, isExists := w.states[v]
		if !isExists {
			continue
		}

		res[v] = value.Clone()
	}

	return res
}

// GetStates returns a copy of the entire states map keyed by field name.
func (w *probeStatesUnit) GetStates() map[string]ProbeUnitResult {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getStates()
}

func (w *probeStatesUnit) getStates() map[string]ProbeUnitResult {
	res := make(map[string]ProbeUnitResult, len(w.states))

	for k, v := range w.states {
		res[k] = v.Clone()
	}

	return res
}

// SetStateMessage updates the Message field of an existing entry or creates a new optional one.
func (w *probeStatesUnit) SetStateMessage(name string, message string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	element, ok := w.states[name]
	if !ok {
		w.states[name] = &ProbeUnitResult{
			Name:    name,
			Message: message,
			Code:    uint16(MessageCodeEverythingOK),

			err:        nil,
			isRequired: false,
		}

		return
	}

	element.Message = message
}

// SetStateError applies the given error to a named entry and recalculates overall status.
func (w *probeStatesUnit) SetStateError(name string, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	value, ok := w.states[name]

	switch {
	case !ok && err != nil:
		w.states[name] = &ProbeUnitResult{
			Name:       name,
			Message:    err.Error(),
			Code:       uint16(w.e.ErrGetCode(err)), //nolint:gosec // ErrGetCode returns non-negative int, safe conversion
			err:        err,
			isRequired: true,
		}

		w.recalculateStatusLocked()

	case !ok && err == nil:
		return

	case ok && err != nil:
		value.Message = err.Error()
		value.Code = uint16(w.e.ErrGetCode(err)) //nolint:gosec // ErrGetCode returns non-negative int, safe conversion
		value.err = err

		w.recalculateStatusLocked()

	case ok && err == nil:
		value.Message = ""
		value.Code = uint16(w.e.ErrGetCode(nil)) //nolint:gosec // ErrGetCode returns 0 for nil, safe conversion
		value.err = nil

		w.recalculateStatusLocked()
	}
}

// recalculateStatusLocked iterates over all stored probe results and sets the unit status to Failed if any required-field still carries an error; otherwise it resets to Success.
func (w *probeStatesUnit) recalculateStatusLocked() {
	for _, r := range w.states {
		if r.isRequired && r.err != nil {
			w.status = ProbeUnitStatusFailed

			return
		}
	}

	w.status = ProbeUnitStatusSuccess
}
