package execute

import "github.com/skupperproject/skupper/test/frame2"

type PodAnnotate struct {
	frame2.Step
}

func (pa PodAnnotate) Run() error {
	return nil
}
