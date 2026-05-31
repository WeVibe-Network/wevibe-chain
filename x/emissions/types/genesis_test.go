package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wevibe-network/wevibe-chain/x/emissions/types"
)

// validatorPool32yr is the full locked validator emission pool (570,000,000 VIBE in uvibe).
const validatorPool32yr = uint64(570_000_000_000_000)

// contributorPool32yr is the locked 32-year contributor pool (320,000,000 VIBE in uvibe).
const contributorPool32yr = uint64(320_000_000_000_000)

// TestDefaultEmissionPool_Seeds32YearSchedule verifies that DefaultEmissionPool()
// seeds the full locked 32-year emission schedule derived from DefaultParams().
func TestDefaultEmissionPool_Seeds32YearSchedule(t *testing.T) {
	pool := types.DefaultEmissionPool()
	require.NotNil(t, pool)

	p := types.DefaultParams()

	// Derived directly from DefaultParams() — single source of truth.
	require.Equal(t, p.DailyMintAmount, pool.DailyMint)
	require.Equal(t, p.OperatorSharePercent, pool.OperatorShare)
	require.Equal(t, p.ValidatorSharePercent, pool.ValidatorShare)
	require.Equal(t, uint64(80), pool.OperatorShare)
	require.Equal(t, uint64(20), pool.ValidatorShare)

	// Full 32-year pools.
	require.Equal(t, validatorPool32yr, pool.ValidatorPoolRemainingUvibe)
	require.Equal(t, uint64(570_000_000_000_000), pool.ValidatorPoolRemainingUvibe)
	require.Equal(t, contributorPool32yr, pool.ContributorPoolRemainingUvibe)
	require.Equal(t, uint64(320_000_000_000_000), pool.ContributorPoolRemainingUvibe)

	// Contributor pool is the annual cap times the 32 schedule years.
	years := p.ScheduleDurationDays / types.EpochsPerYear
	require.Equal(t, uint64(32), years)
	require.Equal(t, p.ContributorAnnualCapUvibe*years, pool.ContributorPoolRemainingUvibe)

	// Counters begin at zero.
	require.Equal(t, uint64(0), pool.TotalSupply)
	require.Equal(t, uint64(0), pool.Epoch)
	require.Equal(t, uint64(0), pool.ContributorRolloverUvibe)
	require.Equal(t, uint64(0), pool.StartEpoch)
	require.Equal(t, uint64(0), pool.TotalEpochsElapsed)

	// Shares sum to 100 so the pool validates.
	require.NoError(t, pool.Validate())
}

// TestDefaultGenesis_UsesDefaultEmissionPool verifies DefaultGenesis embeds the
// full 32-year pool and validates.
func TestDefaultGenesis_UsesDefaultEmissionPool(t *testing.T) {
	g := types.DefaultGenesis()
	require.NotNil(t, g.EmissionPool)
	require.Equal(t, validatorPool32yr, g.EmissionPool.ValidatorPoolRemainingUvibe)
	require.Equal(t, contributorPool32yr, g.EmissionPool.ContributorPoolRemainingUvibe)
	require.NoError(t, g.Validate())
}
