package fuzz

import (
	"bytes"

	"github.com/uluyol/hdrhist"
)

func Fuzz(data []byte) int {
	r := bytes.NewReader(data)
	lr := hdrhist.NewLogReader(r)
	for lr.Scan() {
		_ = lr.Hist()
	}
	_ = lr.Err()
	return 0
}
