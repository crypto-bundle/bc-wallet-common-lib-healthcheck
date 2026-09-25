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
	"context"
	"log/slog"
	"sync"
)

// probeWorkerAsync executes probe units concurrently using goroutines.
// Waits for all goroutines before returning the aggregated result.
// Healthy units do not block unhealthy ones — partial failures are reported.
type probeWorkerAsync struct {
	l              *slog.Logger
	e              errorFormatterService
	units          []probeService
	unitNames      []string
	maxConcurrency int // 0 = unlimited, >0 = cap
	wg             sync.WaitGroup
	mu             sync.RWMutex
}

func newProbeWorkerAsync(logger *slog.Logger, errFmt errorFormatterService) *probeWorkerAsync {
	return &probeWorkerAsync{
		l:              logger,
		e:              errFmt,
		maxConcurrency: 0, // unlimited by default
	}
}

// SetMaxConcurrency limits the number of concurrent goroutines.
// 0 means no limit (default).
func (w *probeWorkerAsync) SetMaxConcurrency(n int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if n < 0 {
		n = 0
	}
	w.maxConcurrency = n
}

// RegisterUnit adds a probe unit for future execution batches.
func (w *probeWorkerAsync) RegisterUnit(name string, unit probeService) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.units = append(w.units, unit)
	w.unitNames = append(w.unitNames, name)
}

// Execute launches all probe units as goroutines and waits for completion.
// Returns aggregated results — unhealthy units do not cancel others.
func (w *probeWorkerAsync) Execute(ctx context.Context) *ProbeExecutionResult {
	w.mu.RLock()
	units := w.units
	names := w.unitNames
	w.mu.RUnlock()

	result := NewProbeExecutionResult(len(units))
	var resultMu sync.Mutex

	// Semaphore-based concurrency limiter
	bound := w.maxConcurrency
	if bound == 0 || bound >= len(units) {
		bound = len(units) // effectively unlimited
	}
	semaphore := make(chan struct{}, bound)

	for i := 0; i < len(units); i++ {
		// Check context before launching
		select {
		case <-ctx.Done():
			resultMu.Lock()
			for j := i; j < len(names); j++ {
				result.AddEntry(names[j], ctx.Err())
			}
			resultMu.Unlock()
			w.wg.Wait()
			return result
		default:
		}

		// Acquire semaphore slot
		select {
		case semaphore <- struct{}{}:
		case <-ctx.Done():
			resultMu.Lock()
			for j := i; j < len(names); j++ {
				result.AddEntry(names[j], ctx.Err())
			}
			resultMu.Unlock()
			w.wg.Wait()
			return result
		}

		w.wg.Add(1)
		go func(idx int) {
			defer w.wg.Done()
			defer func() { <-semaphore }() // release slot

			// Perform health check
			var err error
			if !units[idx].IsReady(ctx) {
				err = ErrProbeNotReady
			} else if !units[idx].IsHealed(ctx) {
				err = ErrProbeUnhealthy
			}

			resultMu.Lock()
			result.AddEntry(names[idx], err)
			resultMu.Unlock()
		}(i)
	}

	w.wg.Wait()
	return result
}
