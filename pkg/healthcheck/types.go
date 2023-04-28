package healthcheck

//go:generate easyjson types.go

import (
	_ "github.com/mailru/easyjson/gen"
)

type ProbeIndex uint8

const (
	StartupProbeIndex ProbeIndex = iota
	RedinessProbeIndex
	LivenessProbeIndex
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

// easyjson:json
type Status struct {
	IsHealed bool    `json:"is_healed"`
	Message  string  `json:"message"`
	Error    *string `json:"error"`
}
