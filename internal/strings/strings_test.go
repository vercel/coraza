// Copyright 2022 Juan Pablo Tosso and the OWASP Coraza contributors
// SPDX-License-Identifier: Apache-2.0

//go:build !tinygo
// +build !tinygo

package strings

import "testing"

func TestRandomStringConcurrency(t *testing.T) {
	// Make sure random strings don't crash under high concurrency.
	for i := 0; i < 5000; i++ {
		go RandomString(10000)
	}
}