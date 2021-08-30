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
	for i := uint(0); i < 11; i++ {
		index := bitmap.First()
		if index < 0 {
			t.Log("cannot find zero value")
		}
		bitmap.Set(uint(index))
	}


	t.Log(bitmap.String())
}

func TestBitMap_Reset(t *testing.T) {
	bitmap := NewBitMap(100)
	for i := 0; i <= 100; i += 10 {
		bitmap.Set(uint(i))
	}

	t.Log(bitmap.String())
	for i := 0; i <= 100; i += 10 {
		bitmap.Reset(uint(i))
	}
	t.Log(bitmap.String())
}