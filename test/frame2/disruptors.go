package frame2

type Disruptor interface {
	DisruptorEnvValue() string
}

// Disruptors that implement the Inspector interface will
// have its Inspect() function called before each step is
// executed.
//
// The disruptor will then be able to analise whether that
// step is of interest for it or not, or even change the
// step's configuration
type Inspector interface {
	Inspect(step *Step, phase *Phase)
}

// PreFinalizerHook will be executed at the end of the
// test, before all other finalizer tasks, such as the
// re-run of validators marked as final
type PreFinalizerHook interface {
	PreFinalizerHook(runner *Run) error
}
