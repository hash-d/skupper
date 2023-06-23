package frame2

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type Monitor interface {
	Executor

	// Starts the monitoring in a different goroutine
	Monitor(*Run) error

	// Result() error
	Report() error
}

type ValidatorConfig struct {
	Name      string
	Validator Validator
	Interval  time.Duration
}

type MonitorResult struct {
	Timestamp time.Time
	Duration  time.Duration
	Step      *Step
	Id        string
	Result    error
}

type DefaultMonitor struct {
	// A list of validators and their names, which cannot conflict
	// with those in ValidatorConfigs
	Validators map[string]Validator

	// The interval between Validator runs.  For ValidatorConfigs, that's
	// the default, if not specified there
	Interval time.Duration

	// Same as Validators, but allow for per-Validator configuration
	ValidatorConfigs []ValidatorConfig

	// This starts as a copy of ValidatorConfigs, to which the Validators
	// are added, with the default configuration
	validatorConfigs []ValidatorConfig

	// Path where the report should be put.  If not set, TODO
	Path string

	Results map[string][]MonitorResult

	Log

	runner *Run

	// This will be based on runner.Ctx, wrapping it with a cancel function
	// that will be called at TearDown
	ctx    context.Context
	finish context.CancelFunc
}

// Result() error
func (d *DefaultMonitor) Report() error {
	for k, validatorResult := range d.Results {
		var count int
		var errors int
		// Change this to a string or struct result, to be consumed and
		// logged elsewhere?
		log.Printf("Results for %v", k)
		for _, r := range validatorResult {
			count += 1
			if r.Result != nil {
				errors += 1
			}
		}
		log.Printf(
			"%d errors (%3.2f%%) errors in %d executions",
			errors,
			float32(errors)/float32(count)*100.0,
			count,
		)
	}
	return nil
}

// This actually only sets up the monitor, internally.  The actual
// execution of the monitor is on Monitor(r).
//
// For any validators configured on this Monitor with an empty
// Logger configuration, the Monitor will replace it by a no-op
// Monitor (log.New(io.Discard, "[M]", 0))
func (m *DefaultMonitor) Execute() error {

	var parentCtx context.Context
	if m.runner == nil {
		parentCtx = context.Background()
	} else {
		parentCtx = ContextOrDefault(m.runner.ctx)
	}
	m.ctx, m.finish = context.WithCancel(parentCtx)

	// TODO Change this by a file destination?
	nlog := log.New(io.Discard, "[M]", 0)
	//nlog = log.New(os.Stderr, "[M]", 0)
	//m.OrSetLogger(nlog)
	m.SetLogger(nlog)

	m.Results = map[string][]MonitorResult{}

	interval := m.Interval
	if interval == 0 {
		interval = time.Second
	}

	// We create a new runner, disconnected from the test, so that monitor
	// runs do not break the test
	monitorRunner := &Run{}

	for k, v := range m.Validators {
		if val, ok := v.(RunDealer); ok {
			val.SetRunner(monitorRunner, MonitorRunner)
		} else {
			panic(fmt.Sprintf(
				"Validator %T on %s is not a RunDealer, and cannot be used as a monitor",
				v,
				m.runner.GetId(),
			))
		}
		OrSetLogger(v, nlog)
		vc := ValidatorConfig{
			Name:      k,
			Validator: v,
			Interval:  interval,
		}
		m.validatorConfigs = append(m.validatorConfigs, vc)
	}

	return nil
}

func (m *DefaultMonitor) Monitor(runner *Run) error {

	m.runner = runner
	for _, vc := range m.validatorConfigs {
		m.goMonitor(vc)
	}

	return nil
}

func (m *DefaultMonitor) goMonitor(vc ValidatorConfig) {
	ctx := m.ctx

	log.Printf("Starting Monitor %+v", vc)

	go func() {
		os.Stdout = nil
		os.Stderr = nil
		// log.SetOutput(io.Discard)
		done := ctx.Done()

		for ctx.Err() == nil {
			start := time.Now()
			// These two lines should not appear on the logs, as m.Log should be
			// discarding its contents.  If it shows up, it's a problem
			m.Log.Printf("v========================== executing ===========v")
			err := vc.Validator.Validate()
			end := time.Now()
			elapsed := end.Sub(start)
			// TODO: this needs sync.Lock, to avoid "fatal error: concurrent map writes"
			m.Results[vc.Name] = append(m.Results[vc.Name], MonitorResult{
				Timestamp: start,
				Duration:  elapsed,
				Result:    err,
			})

			m.Log.Printf("%v", elapsed)
			timeout := time.After(vc.Interval)
			m.Log.Printf("^========================== executed ===========^")
			select {
			case <-done:
			case <-timeout:
			}
		}
		log.Printf("Monitor finished: %#v", vc)
	}()

}

func (m *DefaultMonitor) Teardown() Executor {
	return &Procedure{
		Fn: func() {
			m.finish()
		},
	}
}
