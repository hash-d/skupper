package execute

import "errors"

type Fail struct {
}

func (f Fail) Execute() error {
	return errors.New("execute.Fail failed as requested")
}
