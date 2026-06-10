package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
)

// CanonicalServeBody encodes the exact serve signing preimage:
//
//	wevibe-serve-v1
//	<org_id>
//	<hex(memory_content_hash)>
//	<epoch>
//	<hex(serve_key_pubkey)>
//	<matched_keywords_sorted_joined_by_comma>
//	<hex(nonce)>
//
// Fields are joined with a single '\n' and no trailing newline.
func CanonicalServeBody(orgID string, memoryHash []byte, epoch uint64, serveKeyPubkey []byte, matchedKeywords []string, nonce []byte) []byte {
	keywords := append([]string(nil), matchedKeywords...)
	sort.Strings(keywords)

	lines := []string{
		"wevibe-serve-v1",
		orgID,
		hex.EncodeToString(memoryHash),
		strconv.FormatUint(epoch, 10),
		hex.EncodeToString(serveKeyPubkey),
		strings.Join(keywords, ","),
		hex.EncodeToString(nonce),
	}

	return []byte(strings.Join(lines, "\n"))
}

// CanonicalDenialBody encodes the exact denial signing preimage:
//
//	wevibe-denial-v1
//	<org_id>
//	<hex(memory_hash)>
//	<epoch>
//	<hex(serve_key_pubkey)>
//	<hex(serve_fingerprint)>
//	<hex(nonce)>
//
// Fields are joined with a single '\n' and no trailing newline.
func CanonicalDenialBody(orgID string, memoryHash []byte, epoch uint64, serveKeyPubkey []byte, serveFingerprint []byte, nonce []byte) []byte {
	lines := []string{
		"wevibe-denial-v1",
		orgID,
		hex.EncodeToString(memoryHash),
		strconv.FormatUint(epoch, 10),
		hex.EncodeToString(serveKeyPubkey),
		hex.EncodeToString(serveFingerprint),
		hex.EncodeToString(nonce),
	}

	return []byte(strings.Join(lines, "\n"))
}

// ComputeServeFingerprint returns:
// SHA256(memory_content_hash || serve_key_pubkey || BigEndianUint64(epoch)).
func ComputeServeFingerprint(memoryHash, servePubkey []byte, epoch uint64) []byte {
	epochBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(epochBytes, epoch)

	payload := make([]byte, 0, len(memoryHash)+len(servePubkey)+len(epochBytes))
	payload = append(payload, memoryHash...)
	payload = append(payload, servePubkey...)
	payload = append(payload, epochBytes...)

	sum := sha256.Sum256(payload)
	return append([]byte(nil), sum[:]...)
}

// ComputeDenialFingerprint returns a deterministic dedup key for denials:
// SHA256(CanonicalDenialBody(..., nonce=nil)).
func ComputeDenialFingerprint(orgID string, memoryHash []byte, epoch uint64, servePubkey []byte, serveFingerprint []byte) []byte {
	body := CanonicalDenialBody(orgID, memoryHash, epoch, servePubkey, serveFingerprint, nil)
	sum := sha256.Sum256(body)
	return append([]byte(nil), sum[:]...)
}
