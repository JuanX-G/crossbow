//go:build leaktest

// Copyright 2026 Maciej "juan_em (JuanX-G)" Woźniak
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Or use the bundled copy in the LICENSE file.
package crossbow

import (
	"go.uber.org/goleak"
	"testing"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
