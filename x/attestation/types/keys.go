package types

import (
	"encoding/hex"
	"fmt"
)

const (
	ModuleName = "attestation"
	StoreKey   = "attestation"
)

const (
	SessionHashLen = 32
)

func AttestationKey(orgID string, sessionHash []byte) []byte {
	return []byte("attestation/" + orgID + "/" + ContentHashToHex(sessionHash))
}

func AttestationPrefix(orgID string, epoch uint64) []byte {
	return []byte(fmt.Sprintf("attestation/%s/%d/", orgID, epoch))
}

func AttestationByEpochKey(orgID string, epoch uint64, sessionHash []byte) []byte {
	return []byte(fmt.Sprintf("session_epoch/%s/%d/%s", orgID, epoch, ContentHashToHex(sessionHash)))
}

func AttestationByEpochPrefix(orgID string, epoch uint64) []byte {
	return []byte(fmt.Sprintf("session_epoch/%s/%d/", orgID, epoch))
}

func ContentHashToHex(hash []byte) string {
	return hex.EncodeToString(hash)
}

const ParamsKey = "params"
const RouterKey = ModuleName