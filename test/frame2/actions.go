package frame2

// These are base Executors and Validators that are used by the framework
// itself, so they cannot be placed on packages that depend on the framework

// Calls a function with no args and no return values; use it to cancel
// contexts, for example
type Procedure struct {
	Fn func()
}

func (p *Procedure) Execute() error {
	p.Fn()
	return nil
}
