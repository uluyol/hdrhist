/*
Adapted from github.com/codahale/hdrhistogram

The MIT License (MIT)

Copyright (c) 2014 Coda Hale

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package hdrhist

import (
	"math"
	"testing"
)

func TestCodahaleHighSigFig(t *testing.T) {
	input := []int64{
		459876, 669187, 711612, 816326, 931423, 1033197, 1131895, 2477317,
		3964974, 12718782,
	}

	hist := WithConfig(Config{
		LowestDiscernible: 459876,
		HighestTrackable:  12718782,
		SigFigs:           5,
	})
	for _, sample := range input {
		hist.Record(sample)
	}

	if v, want := hist.PercentileVal(50).Value, int64(1048575); v != want {
		t.Errorf("Median was %v, but expected %v", v, want)
	}
}

func TestCodahaleValueAtQuantile(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  10000000,
		SigFigs:           3,
	})

	for i := 0; i < 1000000; i++ {
		h.Record(int64(i))
	}

	data := []struct {
		q float64
		v int64
	}{
		{q: 50, v: 500223},
		{q: 75, v: 750079},
		{q: 90, v: 900095},
		{q: 95, v: 950271},
		{q: 99, v: 990207},
		{q: 99.9, v: 999423},
		{q: 99.99, v: 999935},
	}

	for _, d := range data {
		if v := h.PercentileVal(d.q).Value; v != d.v {
			t.Errorf("P%v was %v, but expected %v", d.q, v, d.v)
		}
	}
}

func TestCodahaleMean(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  10000000,
		SigFigs:           3,
	})

	for i := 0; i < 1000000; i++ {
		h.Record(int64(i))
	}

	if v, want := h.Mean(), 500000.013312; v != want {
		t.Errorf("Mean was %v, but expected %v", v, want)
	}
}

func TestCodahaleStdDev(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  10000000,
		SigFigs:           3,
	})

	for i := 0; i < 1000000; i++ {
		h.Record(int64(i))
	}

	if v, want := h.Stdev(), 288675.1403682715; v != want {
		t.Errorf("StdDev was %v, but expected %v", v, want)
	}
}

func TestCodahaleTotalCount(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  10000000,
		SigFigs:           3,
	})

	for i := 0; i < 1000000; i++ {
		h.Record(int64(i))
		if v, want := h.TotalCount(), int64(i+1); v != want {
			t.Errorf("TotalCount was %v, but expected %v", v, want)
		}
	}
}

func TestCodahaleMax(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  10000000,
		SigFigs:           3,
	})

	for i := 0; i < 1000000; i++ {
		h.Record(int64(i))
	}

	if v, want := h.Max(), int64(1000447); v != want {
		t.Errorf("Max was %v, but expected %v", v, want)
	}
}

func TestCodahaleClear(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  10000000,
		SigFigs:           3,
	})

	for i := 0; i < 1000000; i++ {
		h.Record(int64(i))
	}

	h.Clear()

	if v, want := h.Max(), int64(0); v != want {
		t.Errorf("Max was %v, but expected %v", v, want)
	}
}

func TestCodahaleMerge(t *testing.T) {
	h1 := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  1000,
		SigFigs:           3,
	})
	h2 := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  1000,
		SigFigs:           3,
	})

	for i := 0; i < 100; i++ {
		h1.Record(int64(i))
	}

	for i := 100; i < 200; i++ {
		h2.Record(int64(i))
	}

	h1.Add(h2)

	if v, want := h1.PercentileVal(50).Value, int64(99); v != want {
		t.Errorf("Median was %v, but expected %v", v, want)
	}
}

func TestCodahaleMin(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  10000000,
		SigFigs:           3,
	})

	for i := 0; i < 1000000; i++ {
		h.Record(int64(i))
	}

	if v, want := h.Min(), int64(0); v != want {
		t.Errorf("Min was %v, but expected %v", v, want)
	}
}

/*
func TestCodahaleByteSize(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  100000,
		SigFigs:           3,
	})

	if v, want := h.EstByteSize(), 65604; v != want {
		t.Errorf("ByteSize was %v, but expected %d", v, want)
	}
}
*/

func TestCodahaleRecordCorrectedValue(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  100000,
		SigFigs:           3,
	})

	h.RecordCorrected(10, 100)

	if v, want := h.PercentileVal(75).Value, int64(10); v != want {
		t.Errorf("Corrected value was %v, but expected %v", v, want)
	}
}

func TestCodahaleRecordCorrectedValueStall(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  100000,
		SigFigs:           3,
	})

	h.RecordCorrected(1000, 100)

	if v, want := h.PercentileVal(75).Value, int64(800); v != want {
		t.Errorf("Corrected value was %v, but expected %v", v, want)
	}
}

func TestCodahaleDistribution(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 8,
		HighestTrackable:  1024,
		SigFigs:           3,
	})

	for i := 0; i < 1024; i++ {
		h.Record(int64(i))
	}

	actual := h.AllVals()
	if len(actual) != 128 {
		t.Fatalf("Number of bars seen was %v, expected was 128", len(actual))
	}
	for _, b := range actual {
		if b.Count != 8 {
			t.Errorf("Count per bar seen was %v, expected was 8", b.Count)
		}
	}
}

func TestCodahaleNaN(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  100000,
		SigFigs:           3,
	})
	if math.IsNaN(h.Mean()) {
		t.Error("mean is NaN")
	}
	if math.IsNaN(h.Stdev()) {
		t.Error("stddev is NaN")
	}
}

func TestCodahaleSignificantFigures(t *testing.T) {
	const sigFigs = 4
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  10,
		SigFigs:           sigFigs,
	})
	if v := h.Config().SigFigs; v != sigFigs {
		t.Errorf("Significant figures was %v, expected %d", v, sigFigs)
	}
}

func TestCodahaleLowestDiscernible(t *testing.T) {
	const minVal = 2
	h := WithConfig(Config{
		LowestDiscernible: minVal,
		HighestTrackable:  10,
		SigFigs:           3,
	})
	if v := h.Config().LowestDiscernible; v != minVal {
		t.Errorf("LowestDiscernible figures was %v, expected %d", v, minVal)
	}
}

func TestCodahaleHighestTrackableValue(t *testing.T) {
	const maxVal = 11
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  maxVal,
		SigFigs:           3,
	})
	if v := h.Config().HighestTrackable; v != maxVal {
		t.Errorf("HighestTrackableValue figures was %v, expected %d", v, maxVal)
	}
}

func BenchmarkHistogramRecordValue(b *testing.B) {
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  10000000,
		SigFigs:           3,
	})
	for i := 0; i < 1000000; i++ {
		h.Record(int64(i))
	}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		h.Record(100)
	}
}

var _benchHist *Hist

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_benchHist = WithConfig(Config{
			LowestDiscernible: 1,
			HighestTrackable:  120000,
			SigFigs:           3,
		}) // this could track 1ms-2min
	}
}

func TestCodahaleUnitMagnitudeOverflow(t *testing.T) {
	panicked := configPanics(Config{
		LowestDiscernible: 0,
		HighestTrackable:  200,
		SigFigs:           4,
	})
	if !panicked {
		t.Error("invalid LowestDiscernible, expected panic")
	}
}

func TestCodahaleSubBucketMaskOverflow(t *testing.T) {
	hist := WithConfig(Config{
		LowestDiscernible: 2e7,
		HighestTrackable:  1e8,
		SigFigs:           5,
	})
	for _, sample := range [...]int64{1e8, 2e7, 3e7} {
		hist.Record(sample)
	}

	for q, want := range map[float64]int64{
		50:    33554431,
		83.33: 33554431,
		83.34: 100663295,
		99:    100663295,
	} {
		if got := hist.PercentileVal(q).Value; got != want {
			t.Errorf("got %d for %fth percentile. want: %d", got, q, want)
		}
	}
}

/*
func TestCodahaleExportImport(t *testing.T) {
	min := int64(1)
	max := int64(10000000)
	sigfigs := 3
	h := hdrhistogram.New(min, max, sigfigs)
	for i := 0; i < 1000000; i++ {
		if err := h.RecordValue(int64(i)); err != nil {
			t.Fatal(err)
		}
	}

	s := h.Export()

	if v := s.LowestTrackableValue; v != min {
		t.Errorf("LowestTrackableValue was %v, but expected %v", v, min)
	}

	if v := s.HighestTrackableValue; v != max {
		t.Errorf("HighestTrackableValue was %v, but expected %v", v, max)
	}

	if v := int(s.SignificantFigures); v != sigfigs {
		t.Errorf("SignificantFigures was %v, but expected %v", v, sigfigs)
	}

	if imported := hdrhistogram.Import(s); !imported.Equals(h) {
		t.Error("Expected Histograms to be equivalent")
	}

}
*/