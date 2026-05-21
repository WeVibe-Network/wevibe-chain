package types

func DefaultParams() Params {
	return Params{
		DefaultMemoryCapPerEpoch: 10000,
		DefaultServeCapPerEpoch:  50000,
	}
}