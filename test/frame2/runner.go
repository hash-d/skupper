package frame2

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

type Run struct {
	T            *testing.T
	Doc          string // TODO this is currently unused (ie, it's not printed)
	savedT       *testing.T
	currentPhase int
	monitors     []*Monitor
	ctx          context.Context
	cancelCtx    context.CancelFunc
}

func (r *Run) GetContext() context.Context {
	if r.ctx == nil {
		r.ctx, r.cancelCtx = context.WithCancel(context.Background())
		r.savedT.Cleanup(r.cancelCtx)
	}
	return r.ctx
}

func (r *Run) CancelContext() {
	r.cancelCtx()
}

func (r *Run) addMonitor(step *Monitor) error {

	r.monitors = append(r.monitors, step)

	return nil
}

func (r *Run) Report() {

	failed := false
	for _, m := range r.monitors {
		err := (*m).Report()
		if err != nil {
			failed = true
		}
	}
	if failed {
		r.savedT.Errorf("At least one monitor failed")
	}

}

type Phase struct {
	Name      string
	Doc       string
	Setup     []Step
	Teardown  []Step
	MainSteps []Step
	//BaseRunner *base.ClusterTestRunnerBase
	teardowns []Executor
	Runner    *Run

	savedRunner *Run
	previousRun bool
	connected   bool

	Log
}

func processStep(t *testing.T, step Step, id string, Log FrameLogger) error {
	// TODO: replace [R] with own logger with Prefix?
	var err error

	if step.SkipWhen {
		Log.Printf("[R] %v step skipped (%s)", id, step.Doc)
		return nil
	}

	if step.Name != "" {
		// For a named test, run or fail, we work the same.  It's up to t to
		// mark it as failed
		_ = t.Run(step.Name, func(t *testing.T) {
			//log.Printf("[R] %v current test: %q", id, t.Name())
			Log.Printf("[R] %v Doc: %v", id, step.Doc)
			processErr := processStep_(t, step, id, Log)
			if processErr != nil {
				// This makes it easier to find the failures in log files
				Log.Printf("[R] test %v - %q failed", id, t.Name())
				// For named tests, we do not return the error up; we
				// just mark it as a failed test
				t.Errorf("test failed: %v", processErr)
			}
			Log.Printf("[R] %v Subtest %q completed", id, t.Name())
		})
	} else {
		// TODO.  Add the step number (like 2.1.3)
		//Log.Printf("[R] %v current test: %q", id, t.Name())
		Log.Printf("[R] %v doc %q", id, step.Doc)
		err = processStep_(t, step, id, Log)
	}
	return err

}
func processStep_(t *testing.T, step Step, id string, Log FrameLogger) error {
	if step.Modify != nil {
		Log.Printf("[R] %v Modifier %T", id, step.Modify)
		var err error
		if phase, ok := step.Modify.(Phase); ok {
			if phase.Runner == nil {
				phase.Runner = &Run{T: t}
			}
			if phase.Name == "" {
				err = phase.runP(fmt.Sprintf("%v.inner", id))
			} else {
				err = phase.runP(id)
			}
		} else {
			err = step.Modify.Execute()
		}
		if err != nil {
			return fmt.Errorf("modify step failed: %w", err)
		}
	}

	subStepList := step.Substeps
	if step.Substep != nil {
		subStepList = append([]*Step{step.Substep}, step.Substeps...)
	}
	for i, subStep := range subStepList {
		_, err := Retry{
			Fn: func() error {
				return processStep(t, *subStep, fmt.Sprintf("%v.sub%d", id, i), Log)
			},
			Options: step.SubstepRetry,
		}.Run()
		if err != nil {
			return fmt.Errorf("substep failed: %w", err)
		}

	}

	validatorList := step.Validators
	if step.Validator != nil {
		validatorList = append([]Validator{step.Validator}, validatorList...)
	}

	if len(validatorList) > 0 {
		fn := func() error {
			someFailure := false
			someSuccess := false
			var lastErr error
			for i, v := range validatorList {
				Log.Printf("[R] %v.v%d Validator %T", id, i, v)
				err := v.Validate()
				if err == nil {
					someSuccess = true
				} else {
					someFailure = true
					lastErr = err
					Log.Printf("[R] %v.v%d Validator %T failed: %v", id, i, v, err)
					if step.ExpectError {
						Log.Printf("[R] (error expected)")
					}
					// Error or not, we do not break or return; we check all
				}
			}
			if step.ExpectError && someSuccess {
				return fmt.Errorf("error expected, but at least one validator passed")
			}
			if !step.ExpectError && someFailure {
				return fmt.Errorf("at least one validator failed.  last error: %w", lastErr)
			}
			return nil
		}

		_, err := Retry{
			Fn:      fn,
			Options: step.ValidatorRetry,
		}.Run()
		return err
	}
	return nil
}

// For a Phase that did not define a Run, this will create a Run
// and set its T accordingly
//
// This is only for the simplest case, when a single phase is
// required.
//
// If the Phase already had a Runner set, it will fail.
func (p *Phase) RunT(t *testing.T) error {
	if p.Runner == nil {
		p.Runner = &Run{
			T: t,
		}
		p.savedRunner = p.Runner
	} else {
		return fmt.Errorf("Phase.RunT configuration error: cannot reset the Runner")
	}
	return p.Run()
}

func (p Phase) Execute() error {
	return p.Run()
}

func (p *Phase) Run() error {
	return p.runP("")
}

func (p *Phase) runP(id string) error {
	var err error

	var idPrefix string
	if id != "" {
		idPrefix = fmt.Sprintf("%v ", id)
	}

	// If a named phase, and within a *testing.T, create a subtest
	if p.Name != "" && p.Runner.T != nil {
		//savedRunner := p.Runner
		ok := p.Runner.T.Run(p.Name, func(t *testing.T) {
			//log.Printf("[R] %vcurrent test: %q", idPrefix, t.Name())
			p.Log.Printf("[R] %vPhase doc: %v", idPrefix, p.Doc)
			p.Runner = &Run{T: t}
			err = p.run(id)
			p.Log.Printf("[R] %vSubtest %q completed", idPrefix, t.Name())
		})

		//p.Runner = savedRunner
		if !ok && err != nil {
			err = errors.New("test returned not-ok, but no errors")
		}
	} else {
		// otherwise, just run it
		//log.Printf("[R] %vcurrent test: %q", idPrefix, p.Runner.T.Name())
		p.Log.Printf("[R] %vPhase doc: %v", idPrefix, p.Doc)
		err = p.run(id)
	}

	if err != nil {
		if p.Runner.T == nil {
			p.Log.Printf("[R] %vPhase error: %v", idPrefix, err)
		}
	}
	return err
}

func (p *Phase) addMonitor(monitor *Monitor) error {
	return p.savedRunner.addMonitor(monitor)

}

func (p *Phase) run(id string) error {

	var idPrefix string
	if id != "" && p.connected {
		idPrefix = fmt.Sprintf("%v.", id)
	}

	// If the phase has no runner, let's create one, without a *testing.T.  This
	// allows the runner to be used disconneced from the testing module.  This
	// way, Actions can be composed using a Phase
	runner := p.Runner
	if runner == nil {
		p.Runner = &Run{}
		p.savedRunner = p.Runner
		runner = p.Runner
	} else {
		p.connected = true
	}

	runner.currentPhase++

	// The Runner is public; we do not want people messing with it in the middle
	// of a Run
	if p.previousRun && p.Runner != p.savedRunner {
		p.Log.Printf("saved: %v, new: %v", p.savedRunner, p.Runner)
		return fmt.Errorf("Phase.Run configuration error: the Runner was changed")

	} else {
		p.savedRunner = p.Runner
	}
	t := runner.T
	//  The testing.T on the Runner is public.  We don't want people messing with
	//  it either.
	if p.previousRun {
		if t != runner.savedT {
			return fmt.Errorf("Phase.Run configuration error: the *testing.T inside the Runner was changed")
		}
	} else {
		p.savedRunner.savedT = t
		p.previousRun = true
	}

	if t != nil {
		t.Cleanup(p.teardown)
	}

	//	if t != nil && p.Name == "" {
	//		t.Fatal("test name must be defined")
	//	}

	//	if t != nil && p.BaseRunner == nil {
	//		p.BaseRunner = &base.ClusterTestRunnerBase{}
	//	}

	if len(p.Setup) > 0 {
		for i, step := range p.Setup {
			if step.Modify != nil {
				if downerStep, ok := step.Modify.(TearDowner); ok {
					tdFunction := downerStep.Teardown()

					if tdFunction != nil {
						p.Log.Printf("[R] %vInstalled auto-teardown for %T", idPrefix, step.Modify)
						p.teardowns = append(p.teardowns, downerStep.Teardown())
					}
				}
			}
			if err := processStep(t, step, fmt.Sprintf("%v%v.set%d", idPrefix, runner.currentPhase, i), &p.Log); err != nil {
				if t != nil {
					t.Fatalf("setup failed: %v", err)
				}
				return err
			}
			if monitorStep, ok := step.Modify.(Monitor); ok {
				p.addMonitor(&monitorStep)
				monitorStep.Monitor(p.savedRunner)
			}
		}
	}

	var savedErr error
	if len(p.MainSteps) > 0 {
		// log.Printf("Starting main steps")
		for i, step := range p.MainSteps {
			if err := processStep(t, step, fmt.Sprintf("%v%v.main%d", idPrefix, runner.currentPhase, i), &p.Log); err != nil {
				savedErr = err
				if t != nil {
					t.Errorf("test failed: %v", err)
				}
				// TODO this should be pluggable
				//p.BaseRunner.DumpTestInfo(p.Name)
				break
			}
			if monitorStep, ok := step.Modify.(Monitor); ok {
				p.addMonitor(&monitorStep)
				monitorStep.Monitor(p.savedRunner)
			}
		}
	}

	if t == nil {
		// If we're not running under testing.T's supervision, we need to run
		// the teardown ourselves.

		// log.Println("Entering teardown phase")
		// log.Printf("Auto tear downs: %#v", p.teardowns)
		p.teardown()
	}
	return savedErr
}

// TODO: thought for later.  Could a user control the order of individual teardowns (automatic
// and explicit) by using different phases?
func (p *Phase) teardown() {
	t := p.Runner.T
	// TODO: if both p.Teardown and p.teardowns were the same interface, this could be
	// a single loop.  Or: leave the individual tear downs to t.Cleanup

	if len(p.Teardown) > 0 {
		p.Log.Printf("Starting teardown")
		// This one runs in normal order, since the user listed them themselves
		for i, step := range p.Teardown {
			if err := processStep(t, step, fmt.Sprintf("down%v", i), &p.Log); err != nil {
				if t == nil {
					p.Log.Printf("Tear down step %d failed: %v", i, err)
				} else {
					t.Errorf("teardown failed: %v", err)
				}
				// We do not return here; we keep going doing whatever
				// teardown we can
			}
		}
	}

	if len(p.teardowns) > 0 {
		// TODO move this to t.Cleanup and make it depend on t != nil?
		// This one runs in reverse order, since they were added by the setup steps
		p.Log.Printf("Starting auto-teardown")
		for i := len(p.teardowns) - 1; i >= 0; i-- {
			td := p.teardowns[i]
			p.Log.Printf("[R] Teardown: %T", td)
			if err := td.Execute(); err != nil {
				if t == nil {
					p.Log.Printf("auto-teardown failed: %v", err)
				} else {
					t.Errorf("auto-teardown failed: %v", err)
				}
				// We do not return here; we keep going doing whatever
				// teardown we can
			}
		}
	}
}
