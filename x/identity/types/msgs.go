package types

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
)

func (m *MsgMigrateIdentity) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.PasskeyPubkey == "" {
		return ErrInvalidPasskeyPubkey
	}
	if m.PasskeySignature == "" {
		return ErrInvalidPasskeySignature
	}

	passkeyPubkey, err := hex.DecodeString(m.PasskeyPubkey)
	if err != nil || len(passkeyPubkey) != ed25519.PublicKeySize {
		return ErrInvalidPasskeyPubkey
	}

	passkeySignature, err := hex.DecodeString(m.PasskeySignature)
	if err != nil || len(passkeySignature) != ed25519.SignatureSize {
		return ErrInvalidPasskeySignature
	}

	return nil
}
