package types

import (
	"encoding/json"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func (gs *GenesisState) MarshalJSON() ([]byte, error) {
	type Alias GenesisState
	return json.Marshal(&struct{ *Alias }{Alias: (*Alias)(gs)})
}

func (gs *GenesisState) UnmarshalJSON(b []byte) error {
	type Alias GenesisState
	aux := &struct{ *Alias }{Alias: (*Alias)(gs)}
	return json.Unmarshal(b, &aux)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterOrg{},
		&MsgAddMember{},
		&MsgRemoveMember{},
		&MsgUpdateParams{},
		&MsgFundTreasury{},
		&MsgWithdrawTreasury{},
		&MsgSetRepTiers{},
		&MsgSetOrgConfig{},
		&MsgGrantTrialAllowance{},
		&MsgUpdateMemberRole{},
		&MsgRotateEpoch{},
		&MsgTransferLeadership{},
		&MsgCloseOrg{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
