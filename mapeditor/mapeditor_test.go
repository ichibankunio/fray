package mapeditor

import "testing"

func TestEncodeTerrainBasePreservesQuarterBlocks(t *testing.T) {
	tests := []struct {
		base float32
		want uint8
	}{
		{base: 0, want: 0},
		{base: 0.25, want: 1},
		{base: 1, want: 4},
		{base: 1.75, want: 7},
	}
	for _, test := range tests {
		if got := encodeTerrainBase(test.base); got != test.want {
			t.Errorf("encodeTerrainBase(%v) = %d, want %d", test.base, got, test.want)
		}
	}
}
