package execute

import "fmt"

type Fail struct {
	Reason string
}

func (f Fail) Execute() error {
	return fmt.Errorf("execute.Fail failed as requested (%q)", f.Reason)
}
