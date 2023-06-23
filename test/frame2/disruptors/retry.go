package disruptors

import (
	"fmt"
	"log"
	"strconv"

	"github.com/skupperproject/skupper/test/frame2"
)

type MinAllows struct {
	MinAllows int
}

func (m MinAllows) DisruptorEnvValue() string {
	return "MIN_ALLOWS"
}

func (m *MinAllows) Inspect(step *frame2.Step, phase *frame2.Phase) {
	if step.ValidatorRetry.Allow < m.MinAllows {
		step.ValidatorRetry.Allow = m.MinAllows
	}
}

func (m *MinAllows) Configure(conf string) error {
	i, err := strconv.Atoi(conf)
	if err != nil {
		return fmt.Errorf("failed to configure MinAllows: %w", err)
	}
	m.MinAllows = i
	log.Printf("Configured disruptor MIN_ALLOWS=%v", i)
	return nil
}
