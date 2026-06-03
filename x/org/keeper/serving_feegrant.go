package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/x/feegrant"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/wevibe-network/wevibe-chain/x/org/types"
)

const (
	// These must match x/serve's registered message type URLs.
	msgTypeURLSubmitServeBatch  = "/wevibe.serve.v1.MsgSubmitServeBatch"
	msgTypeURLSubmitDenialBatch = "/wevibe.serve.v1.MsgSubmitDenialBatch"
)

func (k *Keeper) grantServingFeegrant(ctx context.Context, orgAccountAddr sdk.AccAddress, servingAddr string) error {
	if servingAddr == "" {
		return nil
	}

	servingAddrParsed, err := sdk.AccAddressFromBech32(servingAddr)
	if err != nil {
		return fmt.Errorf("parse serving address: %w", err)
	}

	if k.feegrantKeeper == nil {
		return types.ErrFeegrantUnavailable
	}

	allowance, err := feegrant.NewAllowedMsgAllowance(
		&feegrant.BasicAllowance{},
		[]string{msgTypeURLSubmitServeBatch, msgTypeURLSubmitDenialBatch},
	)
	if err != nil {
		return fmt.Errorf("build serving feegrant allowance: %w", err)
	}

	if err := k.feegrantKeeper.GrantAllowance(ctx, orgAccountAddr, servingAddrParsed, allowance); err != nil {
		return fmt.Errorf("grant serving feegrant: %w", err)
	}

	return nil
}
