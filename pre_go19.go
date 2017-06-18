
// +build !go1.9

package hdrhist

import "github.com/uluyol/hdrhist/internal/bits"

func clz64(v uint64) int {
	return bits.LeadingZeros64(v)
}
