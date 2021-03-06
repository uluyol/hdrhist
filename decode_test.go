package hdrhist

import "testing"

func TestUtil_decodeIntSize(t *testing.T) {
	tests := []struct {
		b   []byte
		ws  int
		v   int64
		err bool
	}{
		{[]byte{0x01, 0x02, 0x03, 0x04}, 4, 0x01020304, false},
		{[]byte{0x01, 0x02, 0x03, 0x04}, 2, 0x0102, false},
		{[]byte{0x01, 0x02, 0x03, 0x04}, 3, 0x010203, false},
		{[]byte{0x01, 0x02, 0x03, 0x04}, 5, 0, true},
		{[]byte{0x01, 0x02, 0x03, 0x04, 0xf1, 0xf2, 0xf3, 0xf4}, 8, 0x01020304f1f2f3f4, false},
		{[]byte{0x01, 0x02, 0x03, 0x04, 0xf1, 0xf2, 0xf3, 0xf4}, 7, 0x01020304f1f2f3, false},
		{[]byte{0x01, 0x02, 0x03, 0x04, 0xf1, 0xf2, 0xf3, 0xf4}, 9, 0, true},
		{[]byte{0x01, 0x02, 0x03, 0x04, 0xf1, 0xf2, 0xf3, 0xf4}, -1, 0, true},
		{[]byte{0x01, 0x02, 0x03, 0x04, 0xf1, 0xf2, 0xf3, 0xf4}, 0, 0, false},
		{nil, 0, 0, false},
	}

	for i, test := range tests {
		v, err := decodeIntSize(test.b, test.ws)
		if test.err {
			if err != nil {
				continue
			}
			t.Errorf("case %d: want error", i)
			continue
		}
		if err != nil {
			t.Errorf("case %d: unexpected error: %v", i, err)
		}
		if v != test.v {
			t.Errorf("case %d: want %d got %d", i, test.v, v)
		}
	}
}
