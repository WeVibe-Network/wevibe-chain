package types

func DefaultParams() Params {
	return Params{
		MaxAttestationsPerEpoch:    10000,
		RequireAttestationForServe: false,
	}
}

func (p *Params) Validate() error {
	return nil
}