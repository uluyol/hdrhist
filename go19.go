
// +build go1.9

package hdrhist

import "math/bits"

func clz64(v int64) int {
	return bits.LeadingZeros(v)
}
