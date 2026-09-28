/*
 * MIT NON-AI License
 *
 * Copyright (c) 2022-2025 Aleksei Kotelnikov(gudron2s@gmail.com)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of the software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions.
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
 *
 * In addition, the following restrictions apply:
 *
 * 1. The Software and any modifications made to it may not be used for the purpose of training or improving machine learning algorithms,
 * including but not limited to artificial intelligence, natural language processing, or data mining. This condition applies to any derivatives,
 * modifications, or updates based on the Software code. Any usage of the Software in an AI-training dataset is considered a breach of this License.
 *
 * 2. The Software may not be included in any dataset used for training or improving machine learning algorithms,
 * including but not limited to artificial intelligence, natural language processing, or data mining.
 *
 * 3. Any person or organization found to be in violation of these restrictions will be subject to legal action and may be held liable
 * for any damages resulting from such use.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
 * DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
 * OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 *
 */

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

// probeStateUnit executes probe units sequentially until all pass or one fails (fail-fast).
type probeStatesUnit struct {
	mu *sync.RWMutex

	mustBeSuccess []string

	status probeUnitStasus
	states map[string]*ProbeUnitResult
}

func NewProbeStateUnit(requiredStateFields ...string) *probeStatesUnit {
	statesMap := make(map[string]*ProbeUnitResult, len(requiredStateFields))

	for _, v := range requiredStateFields {
		statesMap[v] = &ProbeUnitResult{
			Name:    v,
			Message: "",

			err:        nil,
			isRequired: true,
		}
	}

	return &probeStatesUnit{
		mu: &sync.RWMutex{},

		status:        ProbeUnitStatusSuccess,
		mustBeSuccess: requiredStateFields,

		states: statesMap,
	}
}

func (w *probeStatesUnit) GetStatus() *ProbeUnitStatusReport {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return &ProbeUnitStatusReport{
		Status: w.status,
		States: w.getStates(),
	}
}

func (w *probeStatesUnit) GetState(name string) *ProbeUnitResult {
	w.mu.RLock()
	defer w.mu.RUnlock()

	el, ok := w.states[name]
	if !ok {
		return nil
	}

	return el
}

func (w *probeStatesUnit) GetStatesByNames(names ...string) map[string]ProbeUnitResult {
	w.mu.RLock()
	defer w.mu.RUnlock()

	res := make(map[string]ProbeUnitResult, len(names))

	for _, v := range names {
		el, isExists := w.states[v]
		if !isExists {
			continue
		}

		res[v] = el.Clone()
	}

	return res
}

func (w *probeStatesUnit) GetStates() map[string]ProbeUnitResult {
	w.mu.RLock()
	defer w.mu.RUnlock()

	res := make(map[string]ProbeUnitResult, len(w.states))

	for k, v := range w.states {
		res[k] = v.Clone()
	}

	return res
}

func (w *probeStatesUnit) getStates() map[string]ProbeUnitResult {
	res := make(map[string]ProbeUnitResult, len(w.states))

	for k, v := range w.states {
		res[k] = v.Clone()
	}

	return res
}

func (w *probeStatesUnit) SetStateMessage(name string, mesage string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	el, ok := w.states[name]
	if !ok {
		w.states[name] = &ProbeUnitResult{
			Name:    name,
			Message: mesage,

			err:        nil,
			isRequired: false,
		}

		return
	}

	el.Message = mesage
}

func (w *probeStatesUnit) SetStateError(name string, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	el, ok := w.states[name]

	switch true {
	case !ok && err == nil:
		el = &ProbeUnitResult{
			Name:    name,
			Message: err.Error(),

			err:        err,
			isRequired: false,
		}

		w.states[name] = el

	case !ok && err == nil:
		el = &ProbeUnitResult{
			Name:    name,
			Message: "",

			err:        nil,
			isRequired: false,
		}

		w.states[name] = el

	case ok && err != nil:
		el.Message = err.Error()
		el.err = err

	case ok && err == nil:
		el.Message = ""
		el.err = nil
	}
}
