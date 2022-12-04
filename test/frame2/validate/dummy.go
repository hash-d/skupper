package validate

import (
	"log"

	"github.com/skupperproject/skupper/test/frame2"
)

type Dummy struct {
	frame2.Validate
	Results []error
	round   int
}

func (d Dummy) Run() error {
	ret := d.Results[d.round%len(d.Results)]
	d.round++
	log.Printf("Dummy run %d", d.round)
	return ret
}
