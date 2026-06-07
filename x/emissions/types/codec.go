package types

import (
	"encoding/json"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

type GenesisStateJSON struct {
	EmissionPool     *EmissionPool      `json:"emission_pool"`
	DailyEmissions   []*DailyEmission   `json:"daily_emissions"`
	BootstrapCredits []*BootstrapCredit `json:"bootstrap_credits"`
	AsymmetricGates  []*AsymmetricGate  `json:"asymmetric_gates"`
	BootstrapExpiry  uint64             `json:"bootstrap_expiry"`
}

func (g *GenesisState) MarshalJSON() ([]byte, error) {
	return json.Marshal(GenesisStateJSON{
		EmissionPool:     g.EmissionPool,
		DailyEmissions:   g.DailyEmissions,
		BootstrapCredits: g.BootstrapCredits,
		AsymmetricGates:  g.AsymmetricGates,
		BootstrapExpiry:  g.BootstrapExpiry,
	})
}

func (g *GenesisState) UnmarshalJSON(data []byte) error {
	var gj GenesisStateJSON
	if err := json.Unmarshal(data, &gj); err != nil {
		return err
	}
	g.EmissionPool = gj.EmissionPool
	g.DailyEmissions = gj.DailyEmissions
	g.BootstrapCredits = gj.BootstrapCredits
	g.AsymmetricGates = gj.AsymmetricGates
	g.BootstrapExpiry = gj.BootstrapExpiry
	return nil
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgMintDailyEmission{},
		&MsgUpdateParams{},
		&MsgClaimContributorReward{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
