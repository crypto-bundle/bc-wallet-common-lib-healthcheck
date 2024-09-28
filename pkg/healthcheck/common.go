/*
 *
 *
 * MIT NON-AI License
 *
 * Copyright (c) 2022-2024 Aleksei Kotelnikov(gudron2s@gmail.com)
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
	"context"
	"log"
	"log/slog"
	"time"
)

type configService interface {
	IsDebug() bool

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
}

type probeService interface {
	IsHealed(ctx context.Context) bool
}

type probeHttpServer interface {
	AddProbeUnit(unit probeService)
	ListenAndServe(ctx context.Context) error
}

type loggerService interface {
	NewStdLoggerEntry(fields ...any) *log.Logger
	NewStdNamedLoggerEntry(named string, fields ...any) *log.Logger

	NewSlogLoggerEntry(fields ...any) *slog.Logger
	NewSlogNamedLoggerEntry(named string, fields ...any) *slog.Logger
	NewSlogLoggerEntryWithFields(fields ...slog.Attr) *slog.Logger
}

type errorFormatterService interface {
	// ErrorNoWrap function for pseudo-wrapp error, must be used in case of linter warnings...
	ErrorNoWrap(err error) error
	// ErrNoWrap same with ErrorNoWrap function, just alias for ErrorNoWrap, just short function name...
	ErrNoWrap(err error) error
	ErrorOnly(err error, details ...string) error
	Errorf(err error, format string, args ...interface{}) error
	NewError(details ...string) error
	NewErrorf(format string, args ...interface{}) error
}
