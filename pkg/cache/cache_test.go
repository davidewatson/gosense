// Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cache_test

import (
	"reflect"
	"testing"
	"time"

	"experimental/dwat/gosense/pkg/cache"
)

const maxUpdateTime = 1 // Time in seconds; this is short so testing will be quick(er).

// TestUpdateWithTimeout tests that updates panic iff they timeout, errors are
// propagated, and the cache is updated on error.
func TestUpdateWithTimeout(t *testing.T) {
	var testTable = []struct {
		name    string
		update  cache.Update
		timeout int
		deadman bool
		err     error
		panic   bool
	}{
		{name: "instantaneous doesn't panic", update: instantaneous, timeout: maxUpdateTime, deadman: true, err: nil, panic: false},
		{name: "less than a moment panics", update: amoment, timeout: maxUpdateTime - 1, deadman: true, err: nil, panic: true},
		{name: "more than a moment doesn't panic", update: amoment, timeout: maxUpdateTime + 1, deadman: true, err: nil, panic: false},
		{name: "eternity panics", update: eternity, timeout: maxUpdateTime, deadman: true, err: nil, panic: true},
		{name: "errors are propagated", update: ohmygosh, timeout: maxUpdateTime, deadman: true, err: &reflect.ValueError{}, panic: false},
		{name: "errors are propagated even on timeout", update: eternity, timeout: maxUpdateTime, deadman: false, err: &cache.TimeoutError{}, panic: false},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			// Check that we panic when we want to.
			defer func() {
				if r := recover(); (r != nil) != tt.panic {
					t.Errorf("Panic observed %t, expected %t", r != nil, tt.panic)
				}
			}()

			c := cache.NewCache("", tt.update, tt.timeout)
			err := c.UpdateWithTimeout(tt.deadman)

			// Check that the error is as we expect.
			if reflect.TypeOf(err) != reflect.TypeOf(tt.err) {
				t.Errorf("Error observed %v, expected %v", reflect.TypeOf(err), reflect.TypeOf(tt.err))
			}

			// Check that the cache is updated on error.
			if err != nil && string(cache.FormatError(err)) != string(c.Get()) {
				t.Errorf("Cache observed %s, expected %s", c.Get(), cache.FormatError(err))
			}
		})
	}
}

func instantaneous() ([]byte, error) {
	return nil, nil
}

func eternity() ([]byte, error) {
	// TODO(dwat): A bare sleep and return appears to return immediately and set
	// err to nil in the enclosing goroutine context. Why is that?
	//time.Sleep(maxUpdateTime + 42)
	//return nil, nil
	select {}
}

func amoment() ([]byte, error) {
	select {
	case <-time.After(time.Duration(maxUpdateTime) * time.Second):
	}
	return nil, nil
}

func ohmygosh() ([]byte, error) {
	return nil, &reflect.ValueError{}
}
