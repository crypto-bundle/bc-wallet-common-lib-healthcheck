package healthcheck

//go:generate easyjson probe_status.go
// probeResult records the outcome of a single probe unit's health check.
// Name identifies the unit; Err is nil when healthy, non-nil otherwise.

// easyjson:json
type ProbeUnitStatusReport struct {
	Status probeUnitStasus            `json:"name"`
	States map[string]ProbeUnitResult `json:"states"`
}
