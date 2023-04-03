package validate

import (
	"log"

	"github.com/skupperproject/skupper/test/frame2"
)

type Dummy struct {
	Results []error
	round   int

	frame2.Log
}

func (d *Dummy) Validate() error {
	ret := d.Results[d.round%len(d.Results)]
	d.round++
	log.Printf("Dummy run %d", d.round)
	return ret
}
