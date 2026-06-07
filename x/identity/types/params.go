package types

func DefaultParams() Params {
	return Params{
		Active: true,
	}
}

func (p Params) Validate() error {
	return nil
}
