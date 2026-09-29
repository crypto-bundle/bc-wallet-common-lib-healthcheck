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
	"context"
	"time"

	_ "github.com/mailru/easyjson/gen"
)

type probeStatesService interface {
	SetStateMessage(name string, mesage string)
	SetStateError(name string, err error)

	GetStatus() *ProbeUnitStatusReport
}

// probeCheckerConfigService exposes basic configuration for individual probe checker units.
type probeCheckerConfigService interface {
	GetListenAddress() string
	GetHTTPReadTimeout() time.Duration
	GetHTTPWriteTimeout() time.Duration
	GetRequestURL() string
	GetProbeName() string
}

// configService provides per-probe configuration getters derived from environment variables.
type configService interface {
	IsLivenessProbeEnable() bool
	GetLivenessListenAddress() string
	GetLivenessProbeRequestPath() string
	GetLivenessProbeReadTimeout() time.Duration
	GetLivenessProbeWriteTimeout() time.Duration
	GetLivenessProbeListenPort() uint

	IsReadinessProbeEnable() bool
	GetReadinessListenAddress() string
	GetReadinessProbeRequestPath() string
	GetReadinessProbeReadTimeout() time.Duration
	GetReadinessProbeWriteTimeout() time.Duration
	GetReadinessProbeListenPort() uint

	IsStartupProbeEnable() bool
	GetStartupListenAddress() string
	GetStartupProbeRequestPath() string
	GetStartupProbeReadTimeout() time.Duration
	GetStartupProbeWriteTimeout() time.Duration
	GetStartupProbeListenPort() uint

	IsHealthcheckProbeEnable() bool
	GetHealthcheckListenAddress() string
	GetHealthcheckProbeRequestPath() string
	GetHealthcheckProbeReadTimeout() time.Duration
	GetHealthcheckProbeWriteTimeout() time.Duration
	GetHealthcheckProbeListenPort() uint

	IsPongProbeEnable() bool
	GetPingListenAddress() string
	GetPingProbeRequestPath() string
	GetPingProbeReadTimeout() time.Duration
	GetPingProbeWriteTimeout() time.Duration
	GetPingProbeListenPort() uint
}

// probeHTTPServer abstracts listening and serving a probe HTTP server.
type probeHTTPServer interface {
	ListenAndServe(ctx context.Context) error
}

// errorFormatterService wraps lib-errors utilities for structured error handling across
// the healthcheck package. Interfacebloat is acceptable here because all methods serve a distinct role.
type errorFormatterService interface {
	ErrWithCode(err error, code int) error
	NewErrorWithCode(text string, code int) error
	ErrorGetCode(err error) int
	ErrGetCode(err error) int
	ErrorCodeIsOneOf(err error, codes ...int) (int, bool)
	ErrCodeIsOneOf(err error, codes ...int) (int, bool)
	// ErrorNoWrap wraps an error without adding a new stack frame. Used to satisfy linters that flag silent errors.
	ErrorNoWrap(err error) error
	// ErrNoWrap is an alias for ErrorNoWrap with a shorter name.
	ErrNoWrap(err error) error
	ErrorOnly(err error, details ...string) error
	Error(err error, details ...string) error
	Errorf(err error, format string, args ...interface{}) error
	NewError(details ...string) error
	NewErrorf(format string, args ...interface{}) error
}

// baseConfigService detects the runtime environment mode (prod, stage, test, dev, debug, local).
type baseConfigService interface {
	IsProd() bool
	IsStage() bool
	IsTest() bool
	IsDev() bool
	IsDebug() bool
	IsLocal() bool
}
