package execute

import (
	"context"
	"log"
	"time"
)

// Simply waits for the configured duration.  If a context
// is provided and it gets cancelled, the function will
// return earlier.
type Wait struct {
	Delay time.Duration

	Ctx context.Context
}

func (w Wait) Execute() error {
	log.Printf("Waiting for %v", w.Delay)
	ctx := w.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	done := ctx.Done()
	complete := time.After(w.Delay)
	select {
	case <-done:
		log.Printf("Context cancelled with wait incomplete")
	case <-complete:
		log.Printf("Wait complete")
	}
	return nil
}
