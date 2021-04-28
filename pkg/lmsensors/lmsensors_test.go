package lmsensors_test

import (
	"testing"

	// This configures http handlers to serve pprof data at runtime. It also adds ~3 MB
	// to the size of the resulting binary.
	_ "net/http/pprof"

	"experimental/dwat/gosense/pkg/cache"
	"experimental/dwat/gosense/pkg/lmsensors"
)

func Benchmark(b *testing.B) {
	c := cache.NewCache("sensor", lmsensors.Update, 60*60*24*365)
	b.ResetTimer()

	b.Run("c.Get()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c.Get()
		}
	})

	b.Run("c.UpdateWithTimeout()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = c.UpdateWithTimeout(false)
		}
	})
}
