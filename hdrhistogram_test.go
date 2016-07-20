// Adapted from Gil Tene's HdrHistogram for Go
// See https://hdrhistogram.github.io/HdrHistogram/

package hdrhist

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"
)

func configPanics(cfg Config) (panicked bool) {
	defer func() {
		if v := recover(); v != nil {
			panicked = true
		}
	}()
	var h Hist
	h.Init(cfg)
	return false
}

func TestGTHdrHistogramConfigContract(t *testing.T) {
	tests := []struct {
		cfg   Config
		valid bool
	}{
		// Base valid
		{Config{LowestDiscernible: 1, HighestTrackable: 2, SigFigs: 3}, true},

		// Test SigFig values
		{Config{LowestDiscernible: 1, HighestTrackable: 2, SigFigs: -1}, false},
		{Config{LowestDiscernible: 1, HighestTrackable: 2, SigFigs: 0}, true},
		{Config{LowestDiscernible: 1, HighestTrackable: 2, SigFigs: 1}, true},
		{Config{LowestDiscernible: 1, HighestTrackable: 2, SigFigs: 2}, true},
		{Config{LowestDiscernible: 1, HighestTrackable: 2, SigFigs: 3}, true},
		{Config{LowestDiscernible: 1, HighestTrackable: 2, SigFigs: 4}, true},
		{Config{LowestDiscernible: 1, HighestTrackable: 2, SigFigs: 5}, true},
		{Config{LowestDiscernible: 1, HighestTrackable: 2, SigFigs: 6}, false},

		// Test LowestDiscernible
		{Config{LowestDiscernible: 0, HighestTrackable: 2, SigFigs: 3}, false},
		{Config{LowestDiscernible: 1, HighestTrackable: 2, SigFigs: 3}, true},

		// Test HighestTrackable
		{Config{LowestDiscernible: 1, HighestTrackable: 1, SigFigs: 3}, false},
		{Config{LowestDiscernible: 1, HighestTrackable: 2, SigFigs: 3}, true},
		{Config{LowestDiscernible: 2, HighestTrackable: 3, SigFigs: 3}, false},
		{Config{LowestDiscernible: 2, HighestTrackable: 4, SigFigs: 3}, true},
		{Config{LowestDiscernible: 2, HighestTrackable: 5, SigFigs: 3}, true},
		{Config{LowestDiscernible: 12, HighestTrackable: 30, SigFigs: 3}, true},

		// Test AutoResize
		{Config{LowestDiscernible: 1, HighestTrackable: 2, SigFigs: 3, AutoResize: true}, true},
		{Config{LowestDiscernible: 1, HighestTrackable: 1, SigFigs: 3, AutoResize: true}, true},
		{Config{LowestDiscernible: 1, HighestTrackable: 1, SigFigs: 3, AutoResize: false}, false},
	}

	for i, test := range tests {
		valid := !configPanics(test.cfg)
		if valid != test.valid {
			t.Errorf("case %d (%+v), want %t got %t", i, test.cfg, test.valid, valid)
		}
	}
}

func TestGTHdrHistogramEmpty(t *testing.T) {
	h := New(3)
	if v := h.Min(); v != 0 {
		t.Errorf("Min(): want 0 got %f", v)
	}
	if v := h.Max(); v != 0 {
		t.Errorf("Max(): want 0 got %f", v)
	}
	if v := h.Mean(); v != 0 {
		t.Errorf("Mean(): want 0 got %f", v)
	}
	if v := h.Stdev(); v != 0 {
		t.Errorf("Stdev(): want 0 got %f", v)
	}
	if v := h.Val(0).Percentile; v != 100 {
		t.Errorf("Val(0).Percentile: want 100 got %f", v)
	}
}

func verifyMax(h *Hist, t *testing.T) {
	var computedMax int64
	for i := 0; i < len(h.b.counts); i++ {
		if h.b.counts[i] > 0 {
			computedMax = h.b.valueFor(i)
		}
	}
	if computedMax != 0 {
		computedMax = h.b.highestEquiv(computedMax)
	}
	if v := h.Max(); computedMax != v {
		t.Errorf("Max(): want %d, got %d", computedMax, v)
	}
}

func doesPanic(f func()) (b bool) {
	defer func() {
		if v := recover(); v != nil {
			b = true
		}
	}()

	f()
	return false
}

func TestGTHdrHistogramRecord(t *testing.T) {
	const highest = 3600 * 1e6 // 1 hour in μs
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  highest,
		SigFigs:           3,
		AutoResize:        false,
	})

	h.Record(4)
	if v := h.Val(4).Count; v != 1 {
		t.Errorf("Val(4).Count: want %f got %f", 1, v)
	}
	if v := h.TotalCount(); v != 1 {
		t.Errorf("TotalCount(): want %f got %f", 1, v)
	}
	verifyMax(h, t)

	h.Clear()
	if !doesPanic(func() { h.Record(highest * 3) }) {
		t.Error("want panic when recording too high val")
	}
}

func TestGTHdrHistogramLarge(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 20000000,
		HighestTrackable:  100000000,
		SigFigs:           5,
		AutoResize:        false,
	})
	h.Record(100000000)
	h.Record(20000000)
	h.Record(30000000)

	tests := []struct {
		percentile float64
		value      int64
	}{
		{50, 20000000}, {83.33, 30000000}, {83.34, 100000000}, {99, 100000000},
	}
	for _, test := range tests {
		if v := h.PercentileVal(test.percentile).Value; !h.b.areEquiv(v, test.value) {
			t.Errorf("PercentileVal(%f).Value: want %d got %d", test.percentile, test.value, v)
		}
	}
}

func TestGTHdrHistogramClear(t *testing.T) {
	h := New(4)
	h.Record(12)
	h.Clear()
	if c := h.Val(12).Count; c != 0 {
		t.Errorf("Val(12).Count: want 0 got %d", c)
	}
	if c := h.TotalCount(); c != 0 {
		t.Errorf("TotalCount(): want 0 got %d", c)
	}
	verifyMax(h, t)
}

func TestGTHdrHistogramRecordCorrected(t *testing.T) {
	const highest = 3600 * 1e6 // 1 hour in μs
	cfg := Config{
		LowestDiscernible: 1,
		HighestTrackable:  highest,
		SigFigs:           3,
		AutoResize:        false,
	}
	hist := WithConfig(cfg)
	raw := WithConfig(cfg)
	hist.RecordCorrected(4, 1)
	raw.Record(4)
	tests := []struct {
		val    int64
		ccount int64
		rcount int64
	}{
		{1, 1, 0}, {2, 1, 0}, {3, 1, 0}, {4, 1, 1},
	}
	for _, test := range tests {
		if c := hist.Val(test.val).Count; c != test.ccount {
			t.Errorf("hist.Val(%d).Count: want %d got %d", test.val, test.ccount, c)
		}
		if c := raw.Val(test.val).Count; c != test.rcount {
			t.Errorf("raw.Val(%d).Count: want %d got %d", test.val, test.rcount, c)
		}
	}
	if c := hist.TotalCount(); c != 4 {
		t.Errorf("hist.TotalCount(): want 4 got %d", c)
	}

	verifyMax(hist, t)
}

func TestGTHdrHistogramAdd(t *testing.T) {
	const highest = 3600 * 1e6
	cfg := Config{
		LowestDiscernible: 1,
		HighestTrackable:  highest,
		SigFigs:           3,
		AutoResize:        false,
	}
	h := WithConfig(cfg)
	o := WithConfig(cfg)
	h.Record(4)
	h.Record(4000)
	o.Record(4)
	o.Record(4000)
	h.Add(o)
	if c := h.Val(4).Count; c != 2 {
		t.Errorf("Val(4).Count: want 2 got %d", c)
	}
	if c := h.Val(4000).Count; c != 2 {
		t.Errorf("Val(4000).Count: want 2 got %d", c)
	}
	if c := h.TotalCount(); c != 4 {
		t.Errorf("TotalCount(): want 4 got %d", c)
	}

	cfg.HighestTrackable = highest * 2
	big := WithConfig(cfg)
	big.Record(4)
	big.Record(4000)
	big.Record(highest * 2) // overflows h
	big.Add(h)
	if c := big.Val(4).Count; c != 3 {
		t.Errorf("big.Val(4).Count: want 3 got %d", c)
	}
	if c := big.Val(4000).Count; c != 3 {
		t.Errorf("big.Val(4000).Count: want 3 got %d", c)
	}
	if c := big.Val(highest * 2).Count; c != 1 {
		t.Errorf("big.Val(%d).Count: want 1 got %d", highest*2, c)
	}
	if c := big.TotalCount(); c != 7 {
		t.Errorf("big.TotalCount(): want 7 got %d", c)
	}

	if !doesPanic(func() { h.Add(big) }) {
		t.Errorf("h.Add(big): expected panic due to overflow")
	}

	verifyMax(big, t)
}

func TestGTHdrHistogramSubtract(t *testing.T) {
	const highest = 3600 * 1e6
	cfg := Config{
		LowestDiscernible: 1,
		HighestTrackable:  highest,
		SigFigs:           3,
		AutoResize:        false,
	}
	h := WithConfig(cfg)
	o := WithConfig(cfg)
	h.Record(5)
	h.Record(5e4)
	o.Record(5)
	o.Record(5e4)
	h.Add(o)
	h.Add(o)
	if c := h.Val(5).Count; c != 3 {
		t.Errorf("added o twice, want h.Val(5).Count = 3 got %d", c)
	}
	if c := h.Val(5e4).Count; c != 3 {
		t.Errorf("added o twice, want h.Val(5e4).Count = 3 got %d", c)
	}
	if c := h.TotalCount(); c != 6 {
		t.Errorf("added o twice, want h.TotalCount() = 6 got %d", c)
	}
	h.Sub(o)
	if c := h.Val(5).Count; c != 2 {
		t.Errorf("subtracted o once, want h.Val(5).Count = 2 got %d", c)
	}
	if c := h.Val(5e4).Count; c != 2 {
		t.Errorf("subtracted o once, want h.Val(5e4).Count = 2 got %d", c)
	}
	if c := h.TotalCount(); c != 4 {
		t.Errorf("subtracted o once, want h.TotalCount() = 4 got %d", c)
	}
	h.Sub(o)
	if c := h.Val(5).Count; c != 1 {
		t.Errorf("subtracted o twice, want h.Val(5).Count = 1 got %d", c)
	}
	if c := h.Val(5e4).Count; c != 1 {
		t.Errorf("subtracted o twice, want h.Val(5e4).Count = 1 got %d", c)
	}
	if c := h.TotalCount(); c != 2 {
		t.Errorf("subtracted o twice, want h.TotalCount() = 2 got %d", c)
	}
	h.Sub(o)
	if c := h.Val(5).Count; c != 0 {
		t.Errorf("subtracted o three times, want h.Val(5).Count = 0 got %d", c)
	}
	if c := h.Val(5e4).Count; c != 0 {
		t.Errorf("subtracted o three times, want h.Val(5e4).Count = 0 got %d", c)
	}
	if c := h.TotalCount(); c != 0 {
		t.Errorf("subtracted o three times, want h.TotalCount() = 0 got %d", c)
	}
	if !doesPanic(func() { h.Sub(o) }) {
		t.Errorf("should not be able to subtract into negative")
	}

	cfg.HighestTrackable = highest * 3
	big := WithConfig(cfg)
	big.Record(5)
	big.Record(5e4)
	big.Record(highest * 3)
	big.Add(big)
	big.Add(big)

	big.Sub(o)
	if c := big.Val(5).Count; c != 3 {
		t.Errorf("big.Val(5).Count: want 3 got %d", c)
	}
	if c := big.Val(5e4).Count; c != 3 {
		t.Errorf("big.Val(5e4).Count; want 3 got %d", c)
	}
	if c := big.Val(highest * 3).Count; c != 4 {
		t.Errorf("big.Val(%d).Count: want 4 got %d", highest*3, c)
	}
	if c := big.TotalCount(); c != 10 {
		t.Errorf("big.TotalCount(): want 10 got %d", c)
	}

	if !doesPanic(func() { o.Sub(big) }) {
		t.Errorf("should not be able to subtract big from smaller")
	}

	verifyMax(big, t)
}

func TestGTHdrHistogramSizeOfEquivalentValueRange(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  3600 * 1e6,
		SigFigs:           3,
		AutoResize:        false,
	})

	tests := []struct {
		x    int64
		want int64
	}{
		{1, 1}, {2500, 2}, {8191, 4}, {8192, 8}, {1e4, 8},
	}

	for _, test := range tests {
		if s := h.b.sizeOfEquivalentValueRange(test.x); s != test.want {
			t.Errorf("sizeOfEquivalentValueRange(%d) want %d got %d", test.x, test.want, s)
		}
	}

	verifyMax(h, t)
}

func TestGTHdrHistogramScaledSizeOfEquivalentValueRange(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1024,
		HighestTrackable:  3600 * 1e6,
		SigFigs:           3,
		AutoResize:        false,
	})

	tests := []struct {
		x    int64
		want int64
	}{
		{1 * 1024, 1 * 1024}, {2500 * 1024, 2 * 1024}, {8191 * 1024, 4 * 1024},
		{8192 * 1024, 8 * 1024}, {10000 * 1024, 8 * 1024},
	}
	for _, test := range tests {
		if s := h.b.sizeOfEquivalentValueRange(test.x); s != test.want {
			t.Errorf("sizeOfEquivalentValueRange(%d) want %d got %d", test.x, test.want, s)
		}
	}

	verifyMax(h, t)
}

func TestGTHdrHistogramLowestEquivalentValue(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  3600 * 1e6,
		SigFigs:           3,
		AutoResize:        false,
	})
	tests := []struct {
		v    int64
		want int64
	}{
		{10007, 10000}, {10009, 10008},
	}
	for _, test := range tests {
		if eq := h.b.lowestEquiv(test.v); eq != test.want {
			t.Errorf("lowestEquiv(%d) want %d got %d", test.v, test.want, eq)
		}
	}
}

func TestGTHdrHistogramScaledLowestEquivalentValue(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1024,
		HighestTrackable:  3600 * 1e6,
		SigFigs:           3,
		AutoResize:        false,
	})
	tests := []struct {
		v    int64
		want int64
	}{
		{10007 * 1024, 10000 * 1024}, {10009 * 1024, 10008 * 1024},
	}
	for _, test := range tests {
		if eq := h.b.lowestEquiv(test.v); eq != test.want {
			t.Errorf("lowestEquiv(%d) want %d got %d", test.v, test.want, eq)
		}
	}
}

func TestGTHdrHistogramHighestEquivalentValue(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1024,
		HighestTrackable:  3600 * 1e6,
		SigFigs:           3,
		AutoResize:        false,
	})
	tests := []struct {
		v    int64
		want int64
	}{
		{8180 * 1024, 8183*1024 + 1023},
		{8191 * 1024, 8191*1024 + 1023},
		{8193 * 1024, 8199*1024 + 1023},
		{9995 * 1024, 9999*1024 + 1023},
		{10007 * 1024, 10007*1024 + 1023},
		{10008 * 1024, 10015*1024 + 1023},
	}
	for _, test := range tests {
		if eq := h.b.highestEquiv(test.v); eq != test.want {
			t.Errorf("highestEquiv(%d) want %d got %d", test.v, test.want, eq)
		}
	}
}

func TestGTHdrHistogramScaledHighestEquivalentValue(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  3600 * 1e6,
		SigFigs:           3,
		AutoResize:        false,
	})
	tests := []struct {
		v    int64
		want int64
	}{
		{8180, 8183}, {8191, 8191}, {8193, 8199}, {9995, 9999},
		{10007, 10007}, {10008, 10015},
	}
	for _, test := range tests {
		if eq := h.b.highestEquiv(test.v); eq != test.want {
			t.Errorf("highestEquiv(%d) want %d got %d", test.v, test.want, eq)
		}
	}
}

func TestGTHdrHistogramMedianEquivalentValue(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  3600 * 1e6,
		SigFigs:           3,
		AutoResize:        false,
	})
	tests := []struct {
		v    int64
		want int64
	}{
		{4, 4}, {5, 5}, {4000, 4001}, {8000, 8002}, {10007, 10004},
	}
	for _, test := range tests {
		if eq := h.b.medianEquiv(test.v); eq != test.want {
			t.Errorf("medianEquiv(%d) want %d got %d", test.v, test.want, eq)
		}
	}
}

func TestGTHdrHistogramScaledMedianEquivalentValue(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1024,
		HighestTrackable:  3600 * 1e6,
		SigFigs:           3,
		AutoResize:        false,
	})
	tests := []struct {
		v    int64
		want int64
	}{
		{4 * 1024, 4*1024 + 512}, {5 * 1024, 5*1024 + 512}, {4000 * 1024, 4001 * 1024},
		{8e3 * 1024, 8002 * 1024}, {10007 * 1024, 10004 * 1024},
	}
	for _, test := range tests {
		if eq := h.b.medianEquiv(test.v); eq != test.want {
			t.Errorf("medianEquiv(%d) want %d got %d", test.v, test.want, eq)
		}
	}
}

func TestGTHdrHistogramIntervalRecording(t *testing.T) {
	cfg := Config{
		LowestDiscernible: 1,
		HighestTrackable:  3600 * 1e6,
		SigFigs:           3,
		AutoResize:        false,
	}
	h := WithConfig(cfg)
	r1 := NewRecorderWithConfig(cfg)
	r2 := NewRecorderWithConfig(cfg)

	for i := int64(0); i < 10e3; i++ {
		h.Record(3000 * i)
		r1.Record(3000 * i)
		r2.Record(3000 * i)
	}
	h2 := r1.IntervalHist(nil)
	if err := sameHistsNoTime(h, h2); err != nil {
		t.Errorf("h and r1.IntervalHist() differ: %v", err)
	}
	h2 = r2.IntervalHist(h2)
	if err := sameHistsNoTime(h, h2); err != nil {
		t.Errorf("h and r2.IntervalHist() differ: %v", err)
	}

	for i := int64(0); i < 5000; i++ {
		h.Record(3000 * i)
		r1.Record(3000 * i)
		r2.Record(3000 * i)
	}
	h3 := r1.IntervalHist(nil)
	sumHist := h2.Clone()
	sumHist.Add(h3)
	if err := sameHistsNoTime(h, sumHist); err != nil {
		t.Errorf("h and sumHist differ: %v", err)
	}
	_ = r2.IntervalHist(nil)

	for i := int64(5000); i < 10000; i++ {
		h.Record(3000 * i)
		r1.Record(3000 * i)
		r2.Record(3000 * i)
	}
	h4 := r1.IntervalHist(nil)
	h4.Add(h3)
	if err := sameHistsNoTime(h4, h2); err != nil {
		t.Errorf("h2 and h4 differ: %v", err)
	}
	h5 := r2.IntervalHist(h4)
	h5.Add(h3)
	if err := sameHistsNoTime(h2, h5); err != nil {
		t.Errorf("h2 and h5 differ: %v", err)
	}
}

// sameHistsNoTime checks whether h1 and h2 have the same
// contents ignoring start/end times.
func sameHistsNoTime(h1, h2 *Hist) error {
	if h1.totalCount != h2.totalCount {
		return errors.Errorf("total counts don't match: %d vs %d", h1.totalCount, h2.totalCount)
	}
	if h1.cfg != h2.cfg {
		return errors.New("configs don't match")
	}
	if !reflect.DeepEqual(h1.b, h2.b) {
		return errors.New("buckets don't match")
	}
	return nil
}
