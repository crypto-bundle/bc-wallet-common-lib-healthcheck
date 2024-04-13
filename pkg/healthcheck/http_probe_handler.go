package healthcheck

import (
	"net/http"
	"sync"
)

type httpHandler struct {
	mu sync.RWMutex

	probes []probeService
}

func (h *httpHandler) AddProbe(svc probeService) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.probes = append(h.probes, svc)
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var isHealed = true
	var message string
	for i := 0; i != len(h.probes); i++ {
		isHealed = isHealed && h.probes[i].IsHealed(r.Context())
	}

	if isHealed {
		message = AppHealthyMessage
		w.WriteHeader(http.StatusOK)
	} else {
		message = AppUnHealthyMessage
		w.WriteHeader(http.StatusTeapot)
	}

	w.Header().Add("Content-Type", "text/plain")
	_, writeErr := w.Write([]byte(message))
	if writeErr != nil {
		return
	}

	return
}

func newHttpHandler() *httpHandler {
	return &httpHandler{
		probes: nil,
	}
}
