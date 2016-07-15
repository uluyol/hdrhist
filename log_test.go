package hdrhist

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func verifyLogReader(t *testing.T, testi int, test string) {
	f, err := os.Open(filepath.Join("testdata", test+".log"))
	if err != nil {
		t.Fatalf("case %d: unable to open file: %v", testi, err)
	}
	defer f.Close()

	r := NewLogReader(f)
	if r.Scan() != true {
		t.Errorf("case %d: expected data", testi)
	} else {
		h := r.Hist()
		b, err := ioutil.ReadFile(filepath.Join("testdata", test+".ans"))
		if err != nil {
			t.Errorf("case %d: unable to read answers: %v", testi, err)
			return
		}
		lines := strings.Split(string(b), "\n")
		numVals, err := strconv.ParseInt(lines[0], 10, 64)
		if err != nil {
			t.Errorf("case %d: malformed val count: %v", testi, err)
			return
		}
		if h.TotalCount() != numVals {
			t.Errorf("case %d: wrong number of values, want %d got %d", testi, numVals, h.TotalCount())
		}
		seenVals := int64(0)
		for i := int64(1); seenVals < numVals; i++ {
			pair := strings.Split(lines[i], ",")
			v, err := strconv.ParseInt(pair[0], 10, 64)
			if err != nil {
				t.Errorf("case %d: malformed value: %v", testi, err)
				return
			}
			c, err := strconv.ParseInt(pair[1], 10, 64)
			if err != nil {
				t.Errorf("case %d: malformed count: %v", testi, err)
				return
			}
			seenVals += c
			hv := h.Val(v)
			if hv.Count != c {
				t.Errorf("case %d: unable to find %d", testi, v)
			}
		}
	}
	if err := r.Err(); err != nil {
		t.Errorf("case %d: unexpected error: %+v", testi, err)
	}
}

func TestLogReader(t *testing.T) {
	tests := []string{
		"single",
		"single_repeated",
		"single_repeated_multi",
	}

	for i, test := range tests {
		verifyLogReader(t, i, test)
	}
}