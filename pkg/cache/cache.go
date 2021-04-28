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

package cache

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"
)

// Update returns data for the cache.
type Update func() ([]byte, error)

type cacher interface {
	Start()
	Get() []byte
	UpdateWithTimeout(deadman bool) error
}

// Cache maintains a cache coherent []byte respresentation.
type Cache struct {
	Name     string       // Name of the cache
	update   Update       // update generates the data used to populate the cache
	data     atomic.Value // data is an atomically updated []byte representation
	Interval int          // Interval (in seconds) determines how often the cache is refreshed
}

// NewCache allocates and initializes a Cache.
func NewCache(name string, update Update, nSeconds int) *Cache {
	cache := Cache{
		Name:     name,
		update:   update,
		Interval: nSeconds,
	}

	cache.data.Store([]byte(UnknownValue))
	return &cache
}

// Start attempts to create a goroutine to periodically update the cache
// forever. If the first synchronous update fails, then it does not start the
// cache, and returns nil. This has the side effect of allowing caches which
// can not run on a particular machine to stop gracefully. It returns a pointer
// to the cache on success.
func (c *Cache) Start() *Cache {
	if err := c.UpdateWithTimeout(false); err != nil {
		return nil
	}

	ticker := time.NewTicker(time.Duration(c.Interval) * time.Second)

	go func() {
		for range ticker.C {
			// Ignore errors. Since UpdateWithTimeout already succeeded once, we're
			// not going to give up now.
			_ = c.UpdateWithTimeout(true)
		}
	}()

	return c
}

// Get returns a consistent slice of byte(s). Note that the underlying memory
// is not protected and assumed to be immutable.
func (c *Cache) Get() []byte {
	return c.data.Load().([]byte)
}

// UpdateWithTimeout calls update asyncronously with a timeout. If action times
// out a message will be logged and the process will optionally panic to avoid
// leaking goroutines (and possibly recover from bad state). After this
// function returns the cache will have been updated, either with new data,
// or with an error string explaining what happened.
func (c *Cache) UpdateWithTimeout(deadman bool) error {
	var err error
	results := make(chan []byte, 1)
	errs := make(chan error, 1)

	go func() {
		result, err := c.update()
		results <- result
		errs <- err
	}()

	select {
	case err = <-errs:
		if err == nil {
			c.data.Store(<-results)
		} else {
			c.data.Store(FormatError(err))
			log.Printf("Update failed, err: %v.\n", err)
		}
	case <-time.After(time.Duration(c.Interval) * time.Second):
		err = error(&TimeoutError{})
		c.data.Store(FormatError(err))
		log.Printf("Update timed out\n")
		if deadman {
			panic("Deadman switch enabled")
		}
	}

	return err
}

// TimeoutError implements the timeout interface from the net package:
// https://golang.org/pkg/net/#Error
type TimeoutError struct{}

// Error returns the string representing the error.
func (e *TimeoutError) Error() string {
	return "cache: timeout error"
}

// Timeout return true because this error is only for timeouts.
func (e *TimeoutError) Timeout() bool {
	return true
}

const (
	// UnknownValue is the value of the cache before it is started.
	UnknownValue = `{"unknown": "!?"}`
	// ErrorValue is the value of the cache if update fails.
	ErrorValue = `{"err": "%v"}`
)

// FormatError returns a JSON message containing the error.
func FormatError(err error) []byte {
	return []byte(fmt.Sprintf(ErrorValue, err))
}
