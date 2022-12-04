package validate

import (
	"log"
)

type Dummy struct {
	Results []error
	round   int
}

func (d *Dummy) Validate() error {
	ret := d.Results[d.round%len(d.Results)]
	d.round++
	log.Printf("Dummy run %d", d.round)
	return ret
}
