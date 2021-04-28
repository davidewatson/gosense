package main

import (
	"testing"

	"experimental/dwat/gosense/pkg/cache"
	"experimental/dwat/gosense/pkg/lmsensors"
	"experimental/dwat/gosense/pkg/lmsensors/classic"
)

func BenchmarkUpdateWithTimeout(b *testing.B) {
	var c *cache.Cache

	c = cache.NewCache("sensor", lmsensors.Update, 10000000)
	b.ResetTimer()

	b.Run("sensor.UpdateWithTimeout()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = c.UpdateWithTimeout(false)
		}
	})

	c = cache.NewCache("csensor", classic.Update, 10000000)
	b.ResetTimer()

	b.Run("csensor.UpdateWithTimeout()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = c.UpdateWithTimeout(false)
		}
	})
}
