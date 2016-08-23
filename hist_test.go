package hdrhist

import "testing"

func onlySigFigs(v int64, sigfigs int32) int64 {
	var stack [20]int8
	top := 0
	sgn := int64(1)
	if v < 0 {
		sgn = -1
		v = -v
	}
	for v > 0 {
		stack[top] = int8(v % 10)
		top++
		v /= 10
	}
	v = 0
	nfig := int32(0)
	for top > 0 {
		i := top - 1
		top--
		if nfig == sigfigs {
			if stack[i] >= 5 {
				v += 1
			}
		}
		v *= 10
		if nfig < sigfigs {
			v += int64(stack[i])
		}
		nfig++
	}
	return v * sgn
}

func TestAlwaysMeetsSigFigs(t *testing.T) {
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
		{q: 50, v: 500e3},
		{q: 75, v: 750e3},
		{q: 90, v: 900e3},
		{q: 95, v: 950e3},
		{q: 99, v: 990e3},
		{q: 99.9, v: 999e3},
		{q: 99.99, v: 1000e3},
	}

	for _, d := range data {
		if v := onlySigFigs(h.PercentileVal(d.q).Value, 3); v != d.v {
			t.Errorf("P%v was %v, but expected %v", d.q, v, d.v)
		}
	}
}

func TestHistValCumCount(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  1e9,
		SigFigs:           3,
	})

	h.Record(1)
	h.Record(1e4)
	h.Record(1e9)

	tests := []struct {
		v  int64
		cc int64
	}{
		{1, 1},
		{1e4, 2},
		{1e10, 3},
	}
	for _, test := range tests {
		if cc := h.Val(test.v).CumCount; cc != test.cc {
			t.Errorf("val %d: bad cumulative count, want %d got %d", test.v, test.cc, cc)
		}
	}
}

func TestValOOB(t *testing.T) {
	h := WithConfig(Config{
		LowestDiscernible: 1,
		HighestTrackable:  1e9,
		SigFigs:           2,
	})

	h.Record(1)
	h.Record(1e4)
	h.Record(1e9)

	tests := []struct {
		v    int64
		want HistVal
	}{
		{0, HistVal{}},
		{-1, HistVal{Value: -1}},
		{1e11, HistVal{Value: 1e11, Count: 0, CumCount: 3, Percentile: 100}},
	}

	for _, test := range tests {
		if v := h.Val(test.v); v != test.want {
			t.Errorf("val %d: want %#v got %#v", test.v, test.want, v)
		}
	}
}
