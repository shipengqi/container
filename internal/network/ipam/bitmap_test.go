package ipam

import "testing"

func TestBitMap_Set(t *testing.T) {
	bitmap := NewBitMap(100)
	for i := 0; i <= 100; i += 10 {
		bitmap.Set(uint(i))
	}
	for i := 0; i <= 100; i += 10 {
		t.Log(bitmap.Has(uint(i)))
	}

	for i := 1; i <= 100; i += 10 {
		t.Log(bitmap.Has(uint(i)))
	}
	bitmap.SetFirst()
	t.Log(bitmap.String())
}
