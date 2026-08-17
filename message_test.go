// Copyright 2026 Maciej "juan_em (JuanX-G)" Woźniak
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Or use the bundled copy in the LICENSE file.

package crossbow

import (
	"context"
	"testing"
	"time"
)

// Test auto-timestamping of the newContextMessage function
func TestMessageTimestamp(t *testing.T) {
	msg := newContextMessage[int, int](context.Background(), 10, nil, true)
	if msg.Timestamp.IsZero() {
		t.Fatalf("expected a non-zero timestamp, found %s", msg.Timestamp.Format(time.UnixDate))
	}

	msg = newContextMessage[int, int](context.Background(), 10, nil, false)
	if !msg.Timestamp.IsZero() {
		t.Fatalf("expected a zero timestamp, found %s", msg.Timestamp.Format(time.UnixDate))
	}
}
