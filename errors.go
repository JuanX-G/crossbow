package crossbow

import (
	"errors"
	"fmt"
)

// Possible error types when validating server configs [ServerConfig].
//
//go:generate stringer -type=ServerConfigErrorType -trimprefix=Config
type ServerConfigErrorType int

const (
	ConfigErrorWorkersZero     ServerConfigErrorType = iota // Workers cannot be zero, at least one thread must be dedicated to the handler
	ConfigErrorRecoverFuncNil                               // The panic recovery is not allowed to be nil
	ConfigErrorMailboxSizeZero                              // The mailbox must not be zero
)

// Error struct for ServerConfig validation errors;
// msg may be used to hold additional data about
// the error.
type ServerConfigError struct {
	ErrType ServerConfigErrorType
	Msg     string
}

func (s ServerConfigError) Error() string {
	return fmt.Sprintf("Server config error of type: %s, with message: %s, occured.", s.ErrType.String(), s.Msg)
}

// Error returned senders of messages inside the queue if the server is terminated before the message is handled
var ErrServerTerminated = errors.New("server is terminated")
