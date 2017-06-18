
// +build !go.19

package hdrhist

import "github.com/uluyol/hdrhist/internal/bits"

func clz64(v int64) int {
	return bits.LeadingZeros(v)
}
