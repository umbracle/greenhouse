package core

func (p *Project) Test() error {
	if err := p.Compile(); err != nil {
		return err
	}

	return nil
}
