package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

// CanonicalServeBody encodes the exact serve signing preimage:
//
//	wevibe-serve-v2
//	<org_id>
//	<hex(memory_content_hash)>
//	<epoch>
//	<hex(serve_key_pubkey)>
//	<hex(nonce)>
//
// v1→v2 (RECALL-PIVOT joint amendment 2026-07-29): matched keywords left the signed preimage — the keyword gate rejected 24/24 legitimate serves (RECALL-PIVOT-SPEC §3.1 E1).
//
// Fields are joined with a single '\n' and no trailing newline.
func CanonicalServeBody(orgID string, memoryHash []byte, epoch uint64, serveKeyPubkey []byte, nonce []byte) []byte {
	lines := []string{
		"wevibe-serve-v2",
		orgID,
		hex.EncodeToString(memoryHash),
		strconv.FormatUint(epoch, 10),
		hex.EncodeToString(serveKeyPubkey),
		hex.EncodeToString(nonce),
	}

	return []byte(strings.Join(lines, "\n"))
}

// CanonicalEventBody encodes the exact consumer-signed E3/E6/E7/E8 event preimage:
//
//	wevibe-event-v1
//	<type token>
//	<org_id>
//	<hex(memory_content_hash)>
//	<epoch>
//	<hex(signer_pubkey)>
//	<type-specific lines…>
//	<hex(nonce)>
//
// Fields are joined with a single '\n' and no trailing newline.
func CanonicalEventBody(eventType EventType, orgID string, memoryHash []byte, epoch uint64, signerPubkey []byte, nonce []byte, entry *EventEntry) ([]byte, error) {
	if entry == nil {
		return nil, fmt.Errorf("event entry body is required")
	}

	token, specific, err := canonicalEventSpecificLines(eventType, entry)
	if err != nil {
		return nil, err
	}

	lines := []string{
		"wevibe-event-v1",
		token,
		orgID,
		hex.EncodeToString(memoryHash),
		strconv.FormatUint(epoch, 10),
		hex.EncodeToString(signerPubkey),
	}
	lines = append(lines, specific...)
	lines = append(lines, hex.EncodeToString(nonce))

	return []byte(strings.Join(lines, "\n")), nil
}

func canonicalEventSpecificLines(eventType EventType, entry *EventEntry) (string, []string, error) {
	switch eventType {
	case EventType_EVENT_TYPE_OUTCOME:
		body := entry.GetOutcome()
		if body == nil {
			return "", nil, fmt.Errorf("outcome event body is required")
		}
		resolution, err := outcomeResolutionToken(body.Resolution)
		if err != nil {
			return "", nil, err
		}
		source, err := outcomeSourceToken(body.Source)
		if err != nil {
			return "", nil, err
		}
		return "outcome", []string{
			hex.EncodeToString(body.EpisodeRef),
			hex.EncodeToString(body.EvidenceRef),
			hex.EncodeToString(body.ServeRef),
			"resolution=" + resolution,
			"source=" + source,
		}, nil
	case EventType_EVENT_TYPE_VALIDITY_PREDICATE:
		body := entry.GetValidityPredicate()
		if body == nil {
			return "", nil, fmt.Errorf("validity_predicate event body is required")
		}
		result, err := predicateResultToken(body.Result)
		if err != nil {
			return "", nil, err
		}
		return "validity_predicate", []string{
			hex.EncodeToString(body.PredicateId),
			"result=" + result,
			hex.EncodeToString(body.EvidenceRef),
		}, nil
	case EventType_EVENT_TYPE_COST_TO_DISCOVER:
		body := entry.GetCostToDiscover()
		if body == nil {
			return "", nil, fmt.Errorf("cost_to_discover event body is required")
		}
		return "cost_to_discover", []string{
			"cycles=" + strconv.FormatUint(body.Cycles, 10),
			"tool_calls=" + strconv.FormatUint(body.ToolCalls, 10),
			"attempts_to_green=" + strconv.FormatUint(uint64(body.AttemptsToGreen), 10),
			hex.EncodeToString(body.EvidenceRef),
		}, nil
	case EventType_EVENT_TYPE_CONVERGENCE:
		body := entry.GetConvergence()
		if body == nil {
			return "", nil, fmt.Errorf("convergence event body is required")
		}
		return "convergence", []string{hex.EncodeToString(body.ConvergenceRef)}, nil
	default:
		return "", nil, ErrInvalidEventType
	}
}

func predicateResultToken(result PredicateResult) (string, error) {
	switch result {
	case PredicateResult_PREDICATE_RESULT_PASS:
		return "pass", nil
	case PredicateResult_PREDICATE_RESULT_FAIL:
		return "fail", nil
	case PredicateResult_PREDICATE_RESULT_ABSENT:
		return "absent", nil
	default:
		return "", fmt.Errorf("predicate result must be pass, fail, or absent")
	}
}

// outcomeResolutionToken maps the E3 tri-state to its preimage token. An
// unobserved use is recorded as unobserved — silence is not a vote (WO-ATTRIB,
// 2026-08-07). UNSPECIFIED is rejected: an outcome must say what was observed.
func outcomeResolutionToken(resolution OutcomeResolution) (string, error) {
	switch resolution {
	case OutcomeResolution_OUTCOME_RESOLUTION_WORKED:
		return "worked", nil
	case OutcomeResolution_OUTCOME_RESOLUTION_DIDNT_WORK:
		return "didnt_work", nil
	case OutcomeResolution_OUTCOME_RESOLUTION_UNOBSERVED:
		return "unobserved", nil
	default:
		return "", fmt.Errorf("outcome resolution must be worked, didnt_work, or unobserved")
	}
}

func outcomeSourceToken(source OutcomeSource) (string, error) {
	switch source {
	case OutcomeSource_OUTCOME_SOURCE_HARVESTED:
		return "harvested", nil
	case OutcomeSource_OUTCOME_SOURCE_USER:
		return "user", nil
	default:
		return "", fmt.Errorf("outcome source must be harvested or user")
	}
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

// ComputeEventFingerprint returns SHA256(CanonicalEventBody(...)).
func ComputeEventFingerprint(body []byte) []byte {
	sum := sha256.Sum256(body)
	return append([]byte(nil), sum[:]...)
}
