package healthcheck

const (
	ListenAddressTag = "healthcheck_listen_address"
	UnitNameTag      = "healthcheck_unit_name"

	RecoveryErrTag   = "recovery_error"
	RecoveryStackTag = "recovery_stack"
	RecoveryTimeTag  = "recovery_time"

	AppHealthyMessage   = "Ok"
	AppUnHealthyMessage = "Failed"
)
