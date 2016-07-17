package hdrhist

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
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

func TestLogReaderMetadata(t *testing.T) {
	f, err := os.Open("testdata/tstamp.log")
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()
	r := NewLogReader(f)
	if !r.Scan() {
		t.Errorf("want hist, got error: %v", r.Err())
		return
	}
	h := r.Hist()
	if r.Scan() {
		t.Errorf("did not want second hist, got: %v", r.Hist())
		return
	}
	if err := r.Err(); err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if c := h.TotalCount(); c != 3 {
		t.Errorf("total count: want 3, got %d", c)
		return
	}
	if st, ok := h.StartTime(); ok {
		want := unixMillisToTime(15000)
		if !fuzzyEqual(st, want) {
			t.Errorf("start time: want %v, got %v", want, st)
		}
	} else {
		t.Error("want start time")
	}
	if et, ok := h.EndTime(); ok {
		want := unixMillisToTime(16003)
		if !fuzzyEqual(et, want) {
			t.Errorf("end time: want %v, got %v", want, et)
		}
	} else {
		t.Errorf("want end time")
	}
	if c := h.Val(100).Count; c != 1 {
		t.Errorf("count at 100: want 1, got %d", c)
	}
	if c := h.Val(10).Count; c != 1 {
		t.Errorf("count at 10: want 1, got %d", c)
	}
	if c := h.Val(44444).Count; c != 1 {
		t.Errorf("count at 44444: want 1, got %d", c)
	}
}

func unixMillisToTime(t int64) time.Time {
	sec := t / 1e3
	nano := (t % 1e3) * 1e6
	return time.Unix(sec, nano)
}

// fuzzyEqual compares to times, allowing them to
// differ by up to 1 μs.
//
// We need to use it because fractions of a second
// are stored as floats in the log.
func fuzzyEqual(t1, t2 time.Time) bool {
	if t1.Unix() != t2.Unix() {
		return false
	}
	delta := t1.Nanosecond() - t2.Nanosecond()
	if delta < 0 {
		delta = -delta
	}
	return delta < 1e3
}
