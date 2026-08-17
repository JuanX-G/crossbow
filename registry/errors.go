package registry

import "errors"

var (
	ErrKeyTaken       = errors.New("provided key is already taken")
	ErrCancelFuncNil  = errors.New("cancel func for a server can not me nil")
	ErrServerNotFound = errors.New("server not found in the registry")
)
