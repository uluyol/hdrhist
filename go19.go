
// +build go1.9

package hdrhist

import "math/bits"

func clz64(v uint64) int {
	return bits.LeadingZeros64(v)
}
