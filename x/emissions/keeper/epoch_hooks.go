package keeper

import (
	"context"
	"fmt"
	"sort"

	"cosmossdk.io/math"

	orgTypes "github.com/wevibe-network/wevibe-chain/x/org/types"
)

const WeVibeEpochIdentifier = "wevibe_epoch"

func (k *Keeper) AfterEpochEnd(ctx context.Context, epochIdentifier string, epochNumber int64) error {
	if epochIdentifier != WeVibeEpochIdentifier {
		return nil
	}

	epoch := uint64(epochNumber)

	emission, err := k.MintDailyEmission(ctx, epoch)
	if err != nil {
		k.logger.Info("failed to mint daily emission", "epoch", epoch, "error", err)
		return nil
	}

	orgs, err := k.orgKeeper.GetAllOrgs(ctx)
	if err != nil {
		k.logger.Info("failed to get orgs", "error", err)
		return nil
	}

	totalPayouts, orgsProcessed, contributorsPaid := k.ProcessOrgPayouts(ctx, epoch, orgs)

	k.logger.Info("epoch emission processed",
		"epoch", epoch,
		"total_emitted", emission.TotalEmitted,
		"orgs_count", len(orgs),
		"orgs_processed", orgsProcessed,
		"contributors_paid", contributorsPaid,
		"total_payouts", totalPayouts,
	)

	return nil
}

func (k *Keeper) ProcessOrgPayouts(ctx context.Context, epoch uint64, orgs []*orgTypes.Org) (math.Int, int, int) {
	var totalPayouts = math.ZeroInt()
	var orgsProcessed int
	var contributorsPaid int

	for _, org := range orgs {
		orgID := org.OrgID

		orgConfig, err := k.orgKeeper.GetOrgConfig(ctx, orgID)
		if err != nil {
			k.logger.Info("failed to get org config", "org", orgID, "error", err)
			continue
		}
		if !orgConfig.ServeAttestationRequired {
			k.logger.Info("org skip - serve attestation not required", "org", orgID)
			continue
		}

		treasuryBalance, err := k.orgKeeper.GetTreasuryBalanceInt(ctx, orgID)
		if err != nil {
			k.logger.Info("failed to get treasury balance", "org", orgID, "error", err)
			continue
		}
		if treasuryBalance.IsZero() || treasuryBalance.IsNegative() {
			k.logger.Info("org skip - zero or negative treasury", "org", orgID, "balance", treasuryBalance)
			continue
		}

		approvedCountByContributor, err := k.memoryKeeper.GetApprovedCountByContributor(ctx, orgID, epoch)
		if err != nil {
			k.logger.Info("failed to get approved memory counts", "org", orgID, "epoch", epoch, "error", err)
			continue
		}
		if len(approvedCountByContributor) == 0 {
			k.logger.Info("org skip - no approved memories in epoch", "org", orgID, "epoch", epoch)
			continue
		}

		minContributions := orgConfig.MinContributionsPerEpoch
		qualifyingContributors := make([]string, 0, len(approvedCountByContributor))
		for contributorID, approvedCount := range approvedCountByContributor {
			if approvedCount >= minContributions {
				qualifyingContributors = append(qualifyingContributors, contributorID)
			}
		}
		if len(qualifyingContributors) == 0 {
			k.logger.Info("org skip - no qualifying contributors", "org", orgID, "epoch", epoch, "min_contributions", minContributions)
			continue
		}
		sort.Strings(qualifyingContributors)

		repTiers, err := k.orgKeeper.GetRepTiers(ctx, orgID)
		if err != nil {
			k.logger.Info("failed to get rep tiers", "org", orgID, "error", err)
			continue
		}

		orgPayout := math.ZeroInt()
		treasury := treasuryBalance

		for _, contributorID := range qualifyingContributors {
			approvedCount := approvedCountByContributor[contributorID]
			profile, err := k.reputationKeeper.GetContributorProfile(ctx, contributorID, orgID)
			reputation := uint64(0)
			if err == nil && profile != nil {
				reputation = profile.TotalApprovedMemories
			}
			tier, err := getRepTierForContributor(repTiers, reputation)
			if err != nil {
				k.logger.Info("failed to match payout tier", "org", orgID, "contributor", contributorID, "error", err)
				continue
			}

			eligibleContributions := approvedCount
			if tier.MaxContributionsPerEpoch > 0 && eligibleContributions > tier.MaxContributionsPerEpoch {
				eligibleContributions = tier.MaxContributionsPerEpoch
			}

			payoutPerMemory, ok := math.NewIntFromString(tier.PayoutPerMemory)
			if !ok {
				k.logger.Info("failed to parse payout_per_memory", "org", orgID, "contributor", contributorID, "value", tier.PayoutPerMemory)
				continue
			}

			payout := payoutPerMemory.Mul(math.NewIntFromUint64(eligibleContributions))
			if payout.IsZero() {
				continue
			}

			if treasury.LT(payout) {
				k.logger.Info("treasury insufficient - stopping org payout",
					"org", orgID,
					"contributor", contributorID,
					"required", payout,
					"available", treasury,
				)
				break
			}

			if err := k.orgKeeper.DebitTreasury(ctx, orgID, payout); err != nil {
				k.logger.Info("failed to debit treasury", "org", orgID, "contributor", contributorID, "error", err)
				continue
			}

			orgPayout = orgPayout.Add(payout)
			treasury = treasury.Sub(payout)
			contributorsPaid++
		}

		totalPayouts = totalPayouts.Add(orgPayout)
		orgsProcessed++
	}

	return totalPayouts, orgsProcessed, contributorsPaid
}

func getRepTierForContributor(repTiers *orgTypes.RepTierConfig, reputation uint64) (*orgTypes.RepTierRecord, error) {
	if repTiers == nil || len(repTiers.Tiers) == 0 {
		return nil, fmt.Errorf("no rep tiers configured")
	}

	for _, tier := range repTiers.Tiers {
		if reputation >= tier.MinReputation && reputation <= tier.MaxReputation {
			return tier, nil
		}
	}

	return repTiers.Tiers[0], nil
}

func (k *Keeper) BeforeEpochStart(ctx context.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}
