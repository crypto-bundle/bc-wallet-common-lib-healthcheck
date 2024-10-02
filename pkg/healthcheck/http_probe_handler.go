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
	"log/slog"
	"net/http"
	"sync"
)

type httpHandler struct {
	l *slog.Logger

	probes []probeService

	mu sync.RWMutex
}

func (h *httpHandler) AddProbe(svc probeService) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.probes = append(h.probes, svc)
}

func (h *httpHandler) ServeHTTP(respWriter http.ResponseWriter, httpReq *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var isHealed = true

	for i := 0; i != len(h.probes); i++ {
		isHealed = isHealed && h.probes[i].IsHealed(httpReq.Context())
	}

	if isHealed {
		h.writeResponse(respWriter, httpReq, http.StatusOK, AppHealthyMessage)

		return
	}

	h.writeResponse(respWriter, httpReq, http.StatusTeapot, AppUnHealthyMessage)
}

func (h *httpHandler) writeResponse(respWriter http.ResponseWriter,
	_ *http.Request,
	statusCode int,
	message string,
) {
	respWriter.WriteHeader(statusCode)
	respWriter.Header().Add("Content-Type", "text/plain")

	_, writeErr := respWriter.Write([]byte(message))
	if writeErr != nil {
		h.l.Error("unable to write http probe response", writeErr)

		return
	}
}

func newHTTPHandler(logger *slog.Logger) *httpHandler {
	return &httpHandler{
		mu: sync.RWMutex{},

		l:      logger,
		probes: nil,
	}
}
