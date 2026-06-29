package types

import "encoding/json"

type GenesisState struct {
	ServeReceipts     []*ServeReceipt           `json:"serve_receipts"`
	DenialReceipts    []*StoredDenialReceipt    `json:"denial_receipts"`
	EpochStats        []*EpochServeStats        `json:"epoch_stats"`
	ContributorServes []*ContributorEpochServes `json:"contributor_serves"`
}

func NewGenesisState(serveReceipts []*ServeReceipt, denialReceipts []*StoredDenialReceipt, epochStats []*EpochServeStats, contributorServes []*ContributorEpochServes) *GenesisState {
	return &GenesisState{
		ServeReceipts:     serveReceipts,
		DenialReceipts:    denialReceipts,
		EpochStats:        epochStats,
		ContributorServes: contributorServes,
	}
}

func (gs *GenesisState) MarshalJSON() ([]byte, error) {
	type Alias GenesisState
	return json.Marshal(&struct{ *Alias }{Alias: (*Alias)(gs)})
}

func (gs *GenesisState) UnmarshalJSON(b []byte) error {
	type Alias GenesisState
	aux := &struct{ *Alias }{Alias: (*Alias)(gs)}
	return json.Unmarshal(b, &aux)
}
