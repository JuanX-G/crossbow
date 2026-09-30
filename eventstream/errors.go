package eventstream

import (
	"errors"
)

var (
	ErrMailboxEmpty = errors.New("Event mailbox is empty")
	ErrSubscriptionCanceled = errors.New("Subscription canceled")
	ErrEventStreamClosed = errors.New("The event stream is closed")
)
