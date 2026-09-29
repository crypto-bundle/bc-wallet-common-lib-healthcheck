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

// ProbeIndex identifies which healthcheck probe type a value represents.
type ProbeIndex uint8

const (
	// StartupProbeIndex identifies the startup probe type.
	StartupProbeIndex ProbeIndex = iota
	// RedinessProbeIndex identifies the readiness probe type.
	RedinessProbeIndex
	// LivenessProbeIndex identifies the liveness probe type.
	LivenessProbeIndex
	// HealthCheckrobeIndex identifies the general healthcheck probe type.
	HealthCheckrobeIndex
	// PingProbeIndex identifies the ping connectivity check probe type.
	PingProbeIndex
	// ProbeIndexMaxValuePlaceholder is not a healthcheck probe index. It tracks the count
	// of all defined ProbeIndex values and must always remain at the last position of the
	// iota block. Do not reorder it.
	ProbeIndexMaxValuePlaceholder = iota
)

const (
	ProbeNameUnsupported = "unsupported_checker_unit"
	ProbeNameStartup     = "startup_checker_unit"
	ProbeNameRediness    = "rediness_checker_unit"
	ProbeNameLiveness    = "liveness_checker_unit"
)

func (i *ProbeIndex) String() string {
	switch *i {
	case StartupProbeIndex:
		return ProbeNameStartup
	case RedinessProbeIndex:
		return ProbeNameRediness
	case LivenessProbeIndex:
		return ProbeNameLiveness
	default:
		return ProbeNameUnsupported
	}
}

// MessageCode represents a structured message code used within probe state tracking.
type MessageCode uint16

const (
	// MessageCodeEverythingOK indicates that no errors are present.
	MessageCodeEverythingOK MessageCode = 0
)

// probeUnitStatus tracks the overall health status of a probe unit.
type probeUnitStatus uint8

const (
	ProbeUnitStatusFailed probeUnitStatus = iota + 1
	ProbeUnitStatusSuccess
)

const (
	ProbeUnitStatusFailedName  = "probe_failed"
	ProbeUnitStatusSuccessName = "startup_checker_unit"
	ProbeUnitStatusMissingName = "<nil>"
)

func (i probeUnitStatus) String() string {
	switch i {
	case ProbeUnitStatusFailed:
		return ProbeUnitStatusFailedName
	case ProbeUnitStatusSuccess:
		return ProbeUnitStatusSuccessName
	default:
		return ProbeUnitStatusMissingName
	}
}
