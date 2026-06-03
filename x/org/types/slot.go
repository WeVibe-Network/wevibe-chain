package types

import "fmt"

func FormatOrgID(slot uint64) string {
	return fmt.Sprintf("weorg-%d", slot)
}
