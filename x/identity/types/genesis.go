package types

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Aliases: []*StoredIdentityAlias{},
		Params:  DefaultParams(),
	}
}

func (g *GenesisState) Validate() error {
	if err := g.Params.Validate(); err != nil {
		return err
	}

	seen := make(map[string]struct{}, len(g.Aliases))
	for i, alias := range g.Aliases {
		if alias == nil {
			return fmt.Errorf("alias[%d] cannot be nil", i)
		}
		if alias.PasskeyPubkey == "" {
			return ErrInvalidPasskeyPubkey
		}

		passkeyPubkey, err := hex.DecodeString(alias.PasskeyPubkey)
		if err != nil || len(passkeyPubkey) != ed25519.PublicKeySize {
			return ErrInvalidPasskeyPubkey
		}

		if alias.WalletAddress == "" {
			return ErrInvalidWalletAddress
		}

		if _, exists := seen[alias.PasskeyPubkey]; exists {
			return fmt.Errorf("duplicate passkey_pubkey: %s", alias.PasskeyPubkey)
		}
		seen[alias.PasskeyPubkey] = struct{}{}
	}

	return nil
}

func (g *GenesisState) MarshalJSON() ([]byte, error) {
	type GenesisStateAlias GenesisState
	alias := struct {
		GenesisStateAlias
	}{
		GenesisStateAlias: GenesisStateAlias{
			Aliases: g.Aliases,
			Params:  g.Params,
		},
	}
	return json.Marshal(alias)
}

func (g *GenesisState) UnmarshalJSON(bz []byte) error {
	type GenesisStateAlias GenesisState
	var alias struct {
		*GenesisStateAlias
	}
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}
	g.Aliases = alias.Aliases
	g.Params = alias.Params
	return nil
}
