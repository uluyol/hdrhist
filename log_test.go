package hdrhist

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
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

func TestLogReaderV1(t *testing.T) {
	tests := []string{
		"v1_single",
		"v1_single_repeated",
		"v1_single_repeated_multi",
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

func TestLogWriter(t *testing.T) {
	tests := []string{
		"wsingle",
		"wsingle_repeated",
		"wsingle_repeated_multi",
		"wmulti",
		"wmulti_repeated",
	}

	for _, test := range tests {
		verifyLogWriter(t, test)
	}
}

func verifyLogWriter(t *testing.T, test string) {
	p := "testdata/" + test + ".ins"
	f, err := os.Open(p)
	if err != nil {
		t.Errorf("case %s: unable to open file %s", p)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	r := NewRecorder(3)
	var buf bytes.Buffer
	w := NewLogWriter(&buf)
	for {
		var startMillis int64
		var endMillis int64
		if err := parseInt64Line(scanner, "start time", &startMillis); err != nil {
			_, ok := err.(lineNotFoundErr)
			if scanner.Err() == nil && ok {
				// no more data
				break
			}
			t.Errorf("case %s: %v", test, err)
			return
		}
		if err := parseInt64Line(scanner, "end time", &endMillis); err != nil {
			t.Errorf("case %s: %v", test, err)
			return
		}
		for scanner.Scan() {
			if scanner.Text() == "---" {
				h := r.IntervalHist(nil)
				h.SetStartTime(unixMillisToTime(startMillis))
				h.SetEndTime(unixMillisToTime(endMillis))
				if err := w.WriteIntervalHist(h); err != nil {
					t.Errorf("case %s: error writing interval hist: %v", test, err)
					return
				}
				break
			}
			v, err := strconv.ParseInt(scanner.Text(), 10, 64)
			if err != nil {
				t.Errorf("case %s: unable to read value: %v", test, err)
				return
			}
			r.Record(v)
		}
	}
	goldenp := "testdata/" + test + ".golden"
	b, err := ioutil.ReadFile(goldenp)
	if err != nil {
		t.Errorf("case %s: unable to open golden: %v", test, err)
		return
	}
	if err := sameContents(b, buf.Bytes()); err != nil {
		t.Errorf("case %s: different contents\nwant:\n%s\ngot:\n%s\nreason: %v", test, b, buf.Bytes(), err)
	}
}

type lineNotFoundErr struct {
	name string
}

func (e lineNotFoundErr) Error() string {
	return fmt.Sprintf("unable to find %d", e.name)
}

func parseInt64Line(s *bufio.Scanner, name string, dest *int64) error {
	if !s.Scan() {
		return lineNotFoundErr{name}
	}
	v, err := strconv.ParseInt(s.Text(), 10, 64)
	if err != nil {
		return errors.Errorf("unable to parse %s: %v", name, err)
	}
	*dest = v
	return nil
}

func sameContents(b1, b2 []byte) error {
	s1 := bufio.NewScanner(bytes.NewReader(b1))
	s2 := bufio.NewScanner(bytes.NewReader(b2))

	for s1.Scan() && s2.Scan() {
		t1 := s1.Text()
		t2 := s2.Text()
		if len(t1) == len(t2) && len(t1) == 0 {
			continue
		}
		if t1[0] == '#' || t2[0] == '"' {
			if t1 != t2 {
				return errors.Errorf("text lines differ: %s\n%s", t1, t2)
			}
			continue
		}
		i1 := strings.LastIndex(t1, ",")
		i2 := strings.LastIndex(t2, ",")
		if i1 != i2 {
			return errors.New("start of data doesn't match")
		}
		if t1[:i1] != t2[:i1] {
			return errors.Errorf("pre-data text doesn't match:\n%s\n%s", t1[:i1], t2[:i1])
		}
		compressed1, err := base64.StdEncoding.DecodeString(t1[i1+1:])
		if err != nil {
			return errors.Errorf("unable to base64 decode %q: %v", t1[i1+1:], err)
		}
		compressed2, err := base64.StdEncoding.DecodeString(t2[i1+1:])
		if err != nil {
			return errors.Errorf("unable to base64 decode %q: %v", t2[i1+1:], err)
		}
		r1, err := zlib.NewReader(bytes.NewReader(compressed1[8:]))
		if err != nil {
			return errors.Errorf("error creating zlib reader: %v", err)
		}
		r2, err := zlib.NewReader(bytes.NewReader(compressed2[8:]))
		if err != nil {
			return errors.Errorf("error creating zlib reader: %v", err)
		}
		b1, err := ioutil.ReadAll(r1)
		r1.Close()
		if err != nil {
			return errors.Errorf("error reading compressed data: %v", err)
		}
		b2, err := ioutil.ReadAll(r2)
		r2.Close()
		if err != nil {
			return errors.Errorf("error reading compressed data: %v", err)
		}
		const highestStart = 4 * 6
		const highestEnd = 4 * 8
		if bytes.Compare(b1[:highestStart], b2[:highestStart]) != 0 {
			return errors.Errorf("different cookie, length, norm offset, sigfigs, or lowest:\n%v\n%v",
				b1[:highestStart], b2[:highestStart])
		}
		highest1 := binary.BigEndian.Uint64(b1[highestStart:highestEnd])
		highest2 := binary.BigEndian.Uint64(b2[highestStart:highestEnd])
		if highest2 < highest1 {
			return errors.Errorf("source has higher highest trackable: %d vs %d", highest1, highest2)
		}
		if bytes.Compare(b1[highestEnd:], b2[highestEnd:]) != 0 {
			return errors.Errorf("different counts:\n%v\n%v", b1[highestEnd:], b2[highestEnd:])
		}
	}
	if s1.Scan() || s2.Scan() {
		return errors.New("different lengths")
	}
	return nil
}
