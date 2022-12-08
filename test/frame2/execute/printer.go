package execute

import (
	"log"
)

type Print struct {
	Message string // if empty, will simply use "%#v"
	Data    []interface{}
}

func (p Print) Execute() error {

	msg := p.Message
	if msg == "" {
		msg = "%#v"
	}
	log.Printf(msg, p.Data...)

	return nil
}
