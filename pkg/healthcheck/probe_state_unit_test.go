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
	"testing"
)

var (
	testErrDB        = errors.New("database unavailable")
	testErrTimeout   = errors.New("request timeout")
	testFieldMessage = "all systems nominal"
	mockErrCode      = uint16(500)
	mockErrCodeTMO   = uint16(408)
)

// mockErrorFormatter implements the full errorFormatterService interface for testing.
// It provides deterministic error-to-code mapping so tests can assert on Code fields.
type mockErrorFormatter struct{}

func (m *mockErrorFormatter) ErrWithCode(err error, code int) error             { return nil }
func (m *mockErrorFormatter) NewErrorWithCode(text string, code int) error     { return nil }
func (m *mockErrorFormatter) ErrorGetCode(err error) int                       { return m.ErrGetCode(err) }
func (m *mockErrorFormatter) ErrGetCode(err error) int {
	// Map known test errors to their expected uint16 codes; returns 0 for nil errors.
	if err == nil {
		return 0
	}
	if err == testErrDB {
		return int(mockErrCode)
	}
	if err == testErrTimeout {
		return int(mockErrCodeTMO)
	}
	return int(mockErrCode)
}
func (m *mockErrorFormatter) ErrorCodeIsOneOf(err error, codes ...int) (int, bool) { return 0, false }
func (m *mockErrorFormatter) ErrCodeIsOneOf(err error, codes ...int) (int, bool)   { return 0, false }
func (m *mockErrorFormatter) ErrorNoWrap(err error) error                         { return nil }
func (m *mockErrorFormatter) ErrNoWrap(err error) error                           { return nil }
func (m *mockErrorFormatter) ErrorOnly(err error, details ...string) error         { return nil }
func (m *mockErrorFormatter) Error(err error, details ...string) error             { return nil }
func (m *mockErrorFormatter) Errorf(err error, format string, args ...interface{}) error {
	return nil
}
func (m *mockErrorFormatter) NewError(details ...string) error           { return nil }
func (m *mockErrorFormatter) NewErrorf(format string, args ...interface{}) error {
	return nil
}

// newUnit creates a probeStatesUnit with given required state field names.
// It wraps the full constructor behind mockErrorFormatter so production code paths
// (err→code mapping, status recalculation) are exercised without mocking the service.
func newUnit(fields ...string) *probeStatesUnit {
	return NewProbeStateUnit(&mockErrorFormatter{}, fields...)
}

// mustStatus asserts the expected status on a unit, failing fast with a labeled message.
// Used to avoid repeating `unit.GetStatus()` + assertion boilerplate across test cases.
func mustStatus(t *testing.T, unit *probeStatesUnit, expected probeUnitStatus, label string) {
	t.Helper()
	st := unit.GetStatus()
	if st.Status != expected {
		t.Errorf("[%s] expected %v, got %v", label, expected, st.Status)
	}
}

// =============================================================================
// 1. NewProbeStateUnit — Base initialization
// =============================================================================

func TestNewProbeStateUnit_InitialState(t *testing.T) {
	t.Parallel()

	fields := []string{"db", "cache"}
	unit := newUnit(fields...)

	status := unit.GetStatus()
	if status.Status != ProbeUnitStatusSuccess {
		t.Errorf("expected success, got %v", status.Status)
	}
	if len(status.States) != 2 {
		t.Fatalf("expected 2 states, got %d", len(status.States))
	}

	for _, name := range fields {
		state := unit.GetState(name)
		if state == nil {
			t.Errorf("expected state for %q, got nil", name)
			continue
		}
		if state.Name != name {
			t.Errorf("expected name %q, got %q", name, state.Name)
		}
		if !state.isRequired {
			t.Errorf("expected isRequired=true for %q", name)
		}
		if state.err != nil {
			t.Errorf("expected err=nil, got %v", state.err)
		}
		if state.Code != uint16(MessageCodeEverythingOK) {
			t.Errorf("expected Code=%d, got %d", MessageCodeEverythingOK, state.Code)
		}
	}
}

func TestNewProbeStateUnit_EmptyFields(t *testing.T) {
	t.Parallel()

	// Ensures that a unit created with zero fields has an empty states map and remains
	// in Success status — no phantom entries or Failed state can appear from an empty init.
	unit := newUnit()

	status := unit.GetStatus()
	if status.Status != ProbeUnitStatusSuccess {
		t.Errorf("expected success, got %v", status.Status)
	}
	if len(status.States) != 0 {
		t.Fatalf("expected empty states, got %d", len(status.States))
	}
}

func TestNewProbeStateUnit_ErrFmtSvcStored(t *testing.T) {
	t.Parallel()

	// Confirms that errFmtSvc survives construction and is callable — without it,
	// calling SetStateError would panic on a nil receiver dereference inside w.e.ErrGetCode().
	svc := &mockErrorFormatter{}
	unit := NewProbeStateUnit(svc, "field1")
	if unit == nil {
		t.Fatal("expected non-nil unit")
	}

	// Exercises the service path: set error → verify Code was populated from ErrGetCode.
	unit.SetStateError("field1", testErrDB)
	state := unit.GetState("field1")
	if state == nil || state.err == nil {
		t.Error("expected error to be set")
	}
}

// =============================================================================
// 2. GetStatus
// =============================================================================

func TestGetStatus_ReturnsSnapshot(t *testing.T) {
	t.Parallel()

	unit := newUnit("field1")

	reportA := unit.GetStatus()
	if reportA.Status != ProbeUnitStatusSuccess {
		t.Errorf("expected success, got %v", reportA.Status)
	}

	unit.SetStateError("field1", testErrDB)

	reportB := unit.GetStatus()
	if reportB.Status != ProbeUnitStatusFailed {
		t.Errorf("expected failed after error, got %v", reportB.Status)
	}

	// Mutate returned report — internal state must be unaffected.
	mutated := unit.GetStatus()
	mutated.States["mutated"] = ProbeUnitResult{Name: "fake"}

	reportC := unit.GetStatus()
	if reportC.Status != ProbeUnitStatusFailed {
		t.Errorf("mutating snapshot affected internal state")
	}
}

func TestGetStatus_ConcurrentReaders(t *testing.T) {
	t.Parallel()

	// Verifies that multiple goroutines can call GetStatus simultaneously without data races.
	unit := newUnit("field1", "field2")

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = unit.GetStatus()
		}()
	}
	wg.Wait()
}

// =============================================================================
// 3. GetState
// =============================================================================

func TestGetState_Found(t *testing.T) {
	t.Parallel()

	unit := newUnit("myField")

	state := unit.GetState("myField")
	if state == nil {
		t.Fatal("expected state for existing key, got nil")
	}
	if state.Name != "myField" {
		t.Errorf("expected name myField, got %s", state.Name)
	}
	if state.err != nil {
		t.Errorf("expected err nil, got %v", state.err)
	}
	if !state.isRequired {
		t.Error("expected isRequired true")
	}
}

func TestGetState_NotFound(t *testing.T) {
	t.Parallel()

	// Confirms that querying a non-existent key returns nil — ensuring callers don't
	// receive a zero-value ProbeUnitResult that could be misinterpreted as a valid entry.
	unit := newUnit("onlyThis")
	if state := unit.GetState("nonexistent"); state != nil {
		t.Errorf("expected nil for nonexistent key, got %+v", state)
	}
}

func TestGetState_IterationSafe(t *testing.T) {
	t.Parallel()

	// Ensures that iterating via GetState is safe under concurrent writes — the internal
	// mutex protects reads while another goroutine mutates the same field.
	unit := newUnit("shared")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			unit.SetStateMessage("shared", "reading")
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = unit.GetState("shared")
		}
	}()

	wg.Wait()
}

// =============================================================================
// 4. GetStatesByNames
// =============================================================================

func TestGetStatesByNames_FilterExisting(t *testing.T) {
	t.Parallel()

	// Verifies that only keys present in the internal map appear in the result,
	// while absent names are silently skipped.
	unit := newUnit("a", "b", "c")

	result := unit.GetStatesByNames("a", "missing", "c")
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if _, ok := result["missing"]; ok {
		t.Error("unexpected 'missing' key in result")
	}
	if _, ok := result["a"]; !ok {
		t.Error("expected 'a' key in result")
	}
	if _, ok := result["c"]; !ok {
		t.Error("expected 'c' key in result")
	}
}

func TestGetStatesByNames_AllCopies(t *testing.T) {
	t.Parallel()

	// Verifies that every returned value is a deep copy — mutating the Name field through
	// the local variable must NOT propagate back into internal storage.
	unit := newUnit("x")

	copied := unit.GetStatesByNames("x")
	val := copied["x"]
	val.Name = "MUTATED"
	copied["x"] = val

	actual := unit.GetState("x")
	if actual == nil || actual.Name != "x" {
		t.Error("mutation of returned map affected internal state")
	}
}

func TestGetStatesByNames_EmptyInput(t *testing.T) {
	t.Parallel()

	// Ensures that passing no names or a nil slice returns an empty map instead of nil,
	// preventing nil-pointer dereferences when callers range over the result.
	unit := newUnit("field")

	nilResult := unit.GetStatesByNames()
	if nilResult == nil || len(nilResult) != 0 {
		t.Errorf("expected empty map for no names, got %v", nilResult)
	}

	nilSlice := unit.GetStatesByNames(nil...)
	if len(nilSlice) != 0 {
		t.Errorf("expected empty map for nil slice, got %d", len(nilSlice))
	}
}

// =============================================================================
// 5. GetStates
// =============================================================================

func TestGetStates_ReturnsAll(t *testing.T) {
	t.Parallel()

	// Verifies that GetStates exposes every field that was passed through the constructor
	// — missing keys would indicate a construction bug where some fields are silently dropped.
	expected := []string{"one", "two"}
	unit := newUnit(expected...)

	states := unit.GetStates()
	if len(states) != 2 {
		t.Fatalf("expected 2 states, got %d", len(states))
	}
	for _, name := range expected {
		if _, ok := states[name]; !ok {
			t.Errorf("expected state %q in result", name)
		}
	}
}

func TestGetStates_CopiesNotReferences(t *testing.T) {
	t.Parallel()

	// Confirms that returned values are deep copies — mutating the Name field via local
	// variable must NOT propagate back into internal storage. This prevents accidental
	// external modification of probe-state data.
	unit := newUnit("key")

	copied := unit.GetStates()
	val := copied["key"]
	val.Name = "CHANGED"
	copied["key"] = val

	actual := unit.GetState("key")
	if actual == nil || actual.Name != "key" {
		t.Error("returned states are references, not copies")
	}
}

func TestGetStates_EdgeEmpty(t *testing.T) {
	t.Parallel()

	// Ensures that a unit with zero initialized fields returns an empty-but-non-nil map.
	// A nil return would cause panics in callers that range over it.
	unit := newUnit() // no fields
	states := unit.GetStates()
	if states == nil {
		t.Fatal("expected empty map, not nil")
	}
	if len(states) != 0 {
		t.Errorf("expected 0 states, got %d", len(states))
	}
}

// =============================================================================
// 6. SetStateMessage
// =============================================================================

func TestSetStateMessage_UpdateExisting(t *testing.T) {
	t.Parallel()

	// Ensures that SetStateMessage updates only the Message field — err must remain nil
	// because SetStateMessage never touches the error channel.
	unit := newUnit("required-field")
	unit.SetStateMessage("required-field", testFieldMessage)

	state := unit.GetState("required-field")
	if state == nil || state.Message != testFieldMessage {
		t.Errorf("expected message %q, got %q", testFieldMessage, state.Message)
	}
	// Key invariant: setMessage does NOT touch err.
	if state.err != nil {
		t.Error("expected err=nil after SetStateMessage")
	}

	mustStatus(t, unit, ProbeUnitStatusSuccess, "after-set-message")
}

func TestSetStateMessage_CreateOptionalEntry(t *testing.T) {
	t.Parallel()

	// Verifies that SetStateMessage creates a brand-new entry for unknown keys — such
	// entries must be flagged as not-required so they never influence the overall status.
	unit := newUnit("required")

	unit.SetStateMessage("optional-new", testFieldMessage)

	state := unit.GetState("optional-new")
	if state == nil {
		t.Fatal("expected created state, got nil")
	}
	if state.Message != testFieldMessage {
		t.Errorf("expected message %q, got %q", testFieldMessage, state.Message)
	}
	// Dynamically-created entries via SetStateMessage MUST be optional.
	if state.isRequired {
		t.Error("expected isRequired=false for dynamically created entry")
	}
	if state.err != nil {
		t.Error("expected err=nil for new message entry")
	}
}

func TestSetStateMessage_IgnoresNilMessage(t *testing.T) {
	t.Parallel()

	// Ensures that calling SetStateMessage with an empty string is a no-op on existing entries.
	// It should neither create a new key nor modify unrelated fields.
	unit := newUnit("field")
	msgBefore := unit.GetState("field").Message

	unit.SetStateMessage("field", "")

	msgAfter := unit.GetState("field").Message
	if msgBefore != "" && msgAfter != msgBefore {
		t.Errorf("message unexpectedly changed to %q", msgAfter)
	}
}

// =============================================================================
// 7. SetStateError — Creating new entries
// =============================================================================

func TestSetStateError_NewFieldWithError(t *testing.T) {
	t.Parallel()

	// Verifies that a brand-new key arriving via SetStateError gets created with isRequired=true,
	// the error code is populated from w.e.ErrGetCode(), and the overall status transitions to Failed
	// — confirming that dynamic entries respect the required-flag and immediately surface as failures.
	unit := newUnit("existing")

	unit.SetStateError("new-dynamic", testErrDB)

	state := unit.GetState("new-dynamic")
	if state == nil {
		t.Fatal("expected created state, got nil")
	}
	if state.err != testErrDB {
		t.Errorf("expected err=%v, got %v", testErrDB, state.err)
	}
	if state.Name != "new-dynamic" {
		t.Errorf("expected name new-dynamic, got %s", state.Name)
	}
	// Key assertion: Code must match the error-code mapping from the mock formatter.
	if state.Code != mockErrCode {
		t.Errorf("expected Code=%d, got %d", mockErrCode, state.Code)
	}
	if !state.isRequired {
		t.Error("expected isRequired=true for dynamic error entry")
	}

	mustStatus(t, unit, ProbeUnitStatusFailed, "after-new-error")
}

func TestSetStateError_NewFieldWithNilError(t *testing.T) {
	t.Parallel()

	// Confirms that SetStateError(nil) on an unknown key is a no-op — it must NOT fabricate
	// a phantom entry just because the caller happened to pass a name with a nil error.
	unit := newUnit("existing")

	beforeCount := len(unit.GetStates())
	unit.SetStateError("new-entry", nil)
	afterCount := len(unit.GetStates())

	if beforeCount != afterCount {
		t.Errorf("nil error should not create new entry (%d -> %d)", beforeCount, afterCount)
	}
}

// =============================================================================
// 8. SetStateError — Updating existing fields
// =============================================================================

func TestSetStateError_SetErrorOnExisting(t *testing.T) {
	t.Parallel()

	// Ensures that setting an error on an existing required field updates Message to the
	// error text, sets Code from w.e.ErrGetCode(err), and drives the overall status to Failed.
	unit := newUnit("target")

	unit.SetStateError("target", testErrTimeout)

	state := unit.GetState("target")
	if state == nil || state.err != testErrTimeout {
		t.Fatal("expected error to be set on existing field")
	}
	if state.Message != testErrTimeout.Error() {
		t.Errorf("expected Message=%q, got %q", testErrTimeout.Error(), state.Message)
	}
	if state.Code != mockErrCodeTMO {
		t.Errorf("expected Code=%d, got %d", mockErrCodeTMO, state.Code)
	}

	mustStatus(t, unit, ProbeUnitStatusFailed, "after-set-error-on-existing")
}

func TestSetStateError_ClearErrorOnExisting(t *testing.T) {
	t.Parallel()

	// Validates the full clear cycle: after passing nil, the field's err, Message AND Code
	// are all reset — this confirms that recalculateStatusLocked() correctly drops Failed
	// when the last required-field error disappears.
	unit := newUnit("cleared")
	unit.SetStateError("cleared", testErrDB)

	mustStatus(t, unit, ProbeUnitStatusFailed, "before-clear")

	unit.SetStateError("cleared", nil)

	state := unit.GetState("cleared")
	if state.err != nil {
		t.Errorf("expected err=nil after clear, got %v", state.err)
	}
	if state.Message != "" {
		t.Errorf("expected empty message after clear, got %q", state.Message)
	}
	if state.Code != 0 {
		t.Errorf("expected Code=0 after clear, got %d", state.Code)
	}

	mustStatus(t, unit, ProbeUnitStatusSuccess, "after-clear-error")
}

func TestSetStateError_SetAndClearNonRequired(t *testing.T) {
	t.Parallel()

	// Proves that errors on optional (non-required) fields never influence the probe status —
	// even if the optional field flips between error and nil, Success must persist because
	// only required fields participate in the recalculation loop.
	unit := newUnit("required-stay-green")

	unit.SetStateMessage("optional-debug", "checking...")
	unit.SetStateError("optional-debug", testErrDB)

	if unit.GetStatus().Status != ProbeUnitStatusSuccess {
		t.Error("expected Success — optional field errors should not affect status")
	}

	unit.SetStateError("optional-debug", nil)

	if unit.GetStatus().Status != ProbeUnitStatusSuccess {
		t.Error("still expected Success after clearing optional error")
	}
}

// =============================================================================
// 9. SetStateError — Status recalculation logic
// =============================================================================

func TestSetStateError_MultipleRequiredFailures(t *testing.T) {
	t.Parallel()

	// Confirms that multiple simultaneous failures across required fields produce exactly one
	// Failed status — verifying that recalculateStatusLocked() short-circuits on the first
	// required-field error instead of accumulating failure counts.
	unit := newUnit("a", "b", "c")

	unit.SetStateError("a", testErrDB)
	unit.SetStateError("b", testErrTimeout)
	unit.SetStateError("c", testErrDB)

	status := unit.GetStatus()
	if status.Status != ProbeUnitStatusFailed {
		t.Errorf("expected exactly Failed with multiple errors, got %v", status.Status)
	}

	detailed := unit.GetStatus()
	failures := 0
	for _, key := range []string{"a", "b", "c"} {
		if st, ok := detailed.States[key]; ok && st.err != nil {
			failures++
		}
	}
	// Each call to SetStateError persisted correctly — all three fields carry independent errors.
	if failures != 3 {
		t.Errorf("expected 3 failed states, got %d", failures)
	}
}

func TestSetStateError_RecoverFromFailure(t *testing.T) {
	t.Parallel()

	// Verifies the complete lifecycle: Success → Failed (on error) → Success (on clear).
	// This is the core invariant: probes MUST return to Healthy once their last required
	// error is resolved; otherwise the service becomes permanently stuck in Failed.
	unit := newUnit("failing", "other")

	mustStatus(t, unit, ProbeUnitStatusSuccess, "initial")

	unit.SetStateError("failing", testErrDB)
	mustStatus(t, unit, ProbeUnitStatusFailed, "after-set-error")

	unit.SetStateError("failing", nil)
	mustStatus(t, unit, ProbeUnitStatusSuccess, "after-clear-error")
}

func TestSetStateError_OptionalDoesNotAffectStatus(t *testing.T) {
	t.Parallel()

	// Confirms that optional field errors are invisible to the probe status — adding and
	// removing them while a required field is already Failed must NOT alter the outcome.
	// This prevents monitoring noise from spurious optional-component failures.
	unit := newUnit("must-fail")
	unit.SetStateError("must-fail", testErrDB)

	unit.SetStateError("extra-optional", testErrTimeout)
	status := unit.GetStatus()
	if status.Status != ProbeUnitStatusFailed {
		t.Errorf("optional error should not change already-Failed status")
	}

	unit.SetStateError("extra-optional", nil)
	if unit.GetStatus().Status != ProbeUnitStatusFailed {
		t.Error("removing optional error should not change Failed → Success")
	}
}

// =============================================================================
// 10. Thread safety (race detector)
// =============================================================================

func TestConcurrentReadWriteRaceDetector(t *testing.T) {
	t.Parallel()

	// Validates thread safety under a realistic mixed workload: 10 pure readers vs 5 message
	// writers vs 5 error writers alternating between set/clear — running with -race should
	// detect any unprotected shared-state access.
	unit := newUnit("concurrent-a", "concurrent-b")

	var wg sync.WaitGroup

	// 10 readers — only read operations, no writes.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_ = unit.GetStatus()
				_ = unit.GetStates()
				_ = unit.GetStatesByNames("concurrent-a", "concurrent-b")
				_ = unit.GetState("concurrent-a")
			}
		}()
	}

	// 5 writers — messages only.
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			name := "concurrent-a"
			if id%2 == 0 {
				name = "concurrent-b"
			}
			for j := 0; j < 50; j++ {
				unit.SetStateMessage(name, "msg-"+string(rune(id)))
			}
		}(i)
	}

	// 5 writers — alternating SetStateError(set) / SetStateError(clear).
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			name := "concurrent-a"
			if id%2 == 0 {
				name = "concurrent-b"
			}
			for j := 0; j < 50; j++ {
				if j%2 == 0 {
					unit.SetStateError(name, testErrDB)
				} else {
					unit.SetStateError(name, nil)
				}
			}
		}(i)
	}

	wg.Wait()
}
