package execute

import (
	"github.com/skupperproject/skupper/test/frame2"
)

type Print struct {
	Message string // if empty, will simply use "%#v"
	Data    []interface{}

	frame2.Log
	frame2.DefaultRunDealer
}

func (p Print) Execute() error {

	msg := p.Message
	if msg == "" {
		msg = "%#v"
	}
	p.Log.Printf(msg, p.Data...)

	return nil
}

func (p Print) Validate() error {
	return p.Execute()
}
