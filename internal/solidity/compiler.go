package solidity

type Compiler interface {
	Compile(i *Input) (*Output, error)
}

type Input struct {
}

type Output struct {
}
