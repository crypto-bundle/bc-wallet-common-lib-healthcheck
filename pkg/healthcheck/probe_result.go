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
	"encoding/json"
)

// ProbeResult records the outcome of a single probe unit's health check.
// Name identifies the unit; Err is nil when healthy, non-nil otherwise.
type ProbeResult struct {
	Name string
	Err  error
}

// IsHealthy returns true when this individual probe result indicates success.
func (r *ProbeResult) IsHealthy() bool {
	return r.Err == nil
}

// ProbeExecutionResult aggregates results from multiple probe unit checks.
// All entries are stored in insertion order for consistent serialization.
type ProbeExecutionResult struct {
	entries []ProbeResult
}

// NewProbeExecutionResult creates an empty execution result ready to collect entries.
func NewProbeExecutionResult(capacity int) *ProbeExecutionResult {
	if capacity <= 0 {
		return &ProbeExecutionResult{entries: make([]ProbeResult, 0)}
	}
	return &ProbeExecutionResult{entries: make([]ProbeResult, 0, capacity)}
}

// AddEntry appends a single probe result to the aggregate.
func (r *ProbeExecutionResult) AddEntry(name string, err error) {
	r.entries = append(r.entries, ProbeResult{Name: name, Err: err})
}

// Entries returns a copy of the underlying slice for read-only access.
func (r *ProbeExecutionResult) Entries() []ProbeResult {
	copied := make([]ProbeResult, len(r.entries))
	copy(copied, r.entries)
	return copied
}

// EntryCount returns the total number of checked units.
func (r *ProbeExecutionResult) EntryCount() int {
	return len(r.entries)
}

// HealthyCount returns how many units passed their health checks.
func (r *ProbeExecutionResult) HealthyCount() int {
	count := 0
	for _, e := range r.entries {
		if e.IsHealthy() {
			count++
		}
	}
	return count
}

// UnhealthyCount returns how many units failed their health checks.
func (r *ProbeExecutionResult) UnhealthyCount() int {
	count := 0
	for _, e := range r.entries {
		if !e.IsHealthy() {
			count++
		}
	}
	return count
}

// AllHealthy returns true when every checked unit is healthy.
func (r *ProbeExecutionResult) AllHealthy() bool {
	return r.UnhealthyCount() == 0
}

// OverallError returns the first encountered error, or nil if all units are healthy.
func (r *ProbeExecutionResult) OverallError() error {
	for _, e := range r.entries {
		if e.Err != nil {
			return e.Err
		}
	}
	return nil
}

// MarshalJSON provides custom JSON encoding so that empty results render as [] rather than null.
func (r *ProbeExecutionResult) MarshalJSON() ([]byte, error) {
	if r.entries == nil {
		r.entries = make([]ProbeResult, 0)
	}
	return json.Marshal(struct {
		Entries []ProbeResult `json:"entries"`
	}{
		Entries: r.entries,
	})
}

// UnmarshalJSON decodes JSON into the result structure.
func (r *ProbeExecutionResult) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Entries []ProbeResult `json:"entries"`
	}{}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	r.entries = aux.Entries
	if r.entries == nil {
		r.entries = make([]ProbeResult, 0)
	}
	return nil
}
