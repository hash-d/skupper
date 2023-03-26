package frame2

import (
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

	// TODO Change this by a file destination?
	nlog := log.New(io.Discard, "[M]", 0)
	m.OrSetLogger(nlog)

	m.Results = map[string][]MonitorResult{}

	interval := m.Interval
	if interval == 0 {
		interval = time.Second
	}

	for k, v := range m.Validators {
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
	ctx := m.runner.GetContext()

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
	}()

}
