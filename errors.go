package crossbow


import ("errors")

// Possible error types when validating server configs [ServerConfig].
type ServerConfigErrorType int
const (
	ConfigErrorThreadsCfgZero ServerConfigErrorType = iota // ThreadsCfg cannot be zero, at least one thread must be dedicated to the handler
	ConfigErrorRecoverFuncNil // The panic recovery is not allowed to be nil
	ConfigErrorMailboxSizeZero // The mailbox must not be zero
)

// Error struct for ServerConfig validation errors;
// msg may be used to hold additional data about
// the error.
type ServerConfigError struct {
	errType ServerConfigErrorType
	msg string
}

func (s ServerConfigError) Error() string {
	switch s.errType {
	case ConfigErrorThreadsCfgZero:
		return "ThreadsCfg cannot be 0"
	case ConfigErrorRecoverFuncNil:
		return "ReoverFn cannot be nil"
	case ConfigErrorMailboxSizeZero:
		return "MailboxSize cannot be 0"
	default:
		return "Unknown config error"
	}
}

// Error returned senders of messages inside the queue if the server is terminated before the message is handled
var ErrServerTerminated = errors.New("server is terminated")
