package execute

type Success struct {
}

func (f Success) Execute() error {
	return nil
}
