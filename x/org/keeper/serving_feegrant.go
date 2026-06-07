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

	// These must match x/org and x/memory registered message type URLs
	// (verified against x/org/types/tx.pb.go and x/memory/types/tx.pb.go).
	msgTypeURLOrgAddMember            = "/wevibe.org.v1.MsgAddMember"
	msgTypeURLOrgRemoveMember         = "/wevibe.org.v1.MsgRemoveMember"
	msgTypeURLOrgUpdateMemberRole     = "/wevibe.org.v1.MsgUpdateMemberRole"
	msgTypeURLOrgSetOrgConfig         = "/wevibe.org.v1.MsgSetOrgConfig"
	msgTypeURLOrgSetServingKey        = "/wevibe.org.v1.MsgSetServingKey"
	msgTypeURLOrgSetServingInfo       = "/wevibe.org.v1.MsgSetServingInfo"
	msgTypeURLOrgSetExtractionProfile = "/wevibe.org.v1.MsgSetExtractionProfile"
	msgTypeURLOrgRotateEpoch          = "/wevibe.org.v1.MsgRotateEpoch"
	msgTypeURLOrgTransferLeadership   = "/wevibe.org.v1.MsgTransferLeadership"
	msgTypeURLOrgCloseOrg             = "/wevibe.org.v1.MsgCloseOrg"
	msgTypeURLOrgGrantTrialAllowance  = "/wevibe.org.v1.MsgGrantTrialAllowance"
	msgTypeURLMemorySubmitCommitment  = "/wevibe.memory.v1.MsgSubmitCommitment"
	msgTypeURLMemoryApproveMemory     = "/wevibe.memory.v1.MsgApproveMemory"
	msgTypeURLMemoryReportMemory      = "/wevibe.memory.v1.MsgReportMemory"
)

var leaderAllowedMsgTypeURLs = []string{
	msgTypeURLOrgAddMember,
	msgTypeURLOrgRemoveMember,
	msgTypeURLOrgUpdateMemberRole,
	msgTypeURLOrgSetOrgConfig,
	msgTypeURLOrgSetServingKey,
	msgTypeURLOrgSetServingInfo,
	msgTypeURLOrgSetExtractionProfile,
	msgTypeURLOrgRotateEpoch,
	msgTypeURLOrgTransferLeadership,
	msgTypeURLOrgCloseOrg,
	msgTypeURLOrgGrantTrialAllowance,
	msgTypeURLMemorySubmitCommitment,
	msgTypeURLMemoryApproveMemory,
	msgTypeURLMemoryReportMemory,
}

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

func (k *Keeper) grantLeaderFeegrant(ctx context.Context, granter sdk.AccAddress, leaderWalletBech32 string) error {
	if leaderWalletBech32 == "" {
		return nil
	}

	leaderWalletAddr, err := sdk.AccAddressFromBech32(leaderWalletBech32)
	if err != nil {
		return fmt.Errorf("parse leader wallet address: %w", err)
	}

	if k.feegrantKeeper == nil {
		return types.ErrFeegrantUnavailable
	}

	allowance, err := feegrant.NewAllowedMsgAllowance(
		&feegrant.BasicAllowance{},
		leaderAllowedMsgTypeURLs,
	)
	if err != nil {
		return fmt.Errorf("build leader feegrant allowance: %w", err)
	}

	if err := k.feegrantKeeper.GrantAllowance(ctx, granter, leaderWalletAddr, allowance); err != nil {
		return fmt.Errorf("grant leader feegrant: %w", err)
	}

	return nil
}
