package ipam

import (
	"fmt"
	"strings"
)

// [1 2 4 8 16 32 64 128]
var bitmask = []byte{1, 1 << 1, 1 << 2, 1 << 3, 1 << 4, 1 << 5, 1 << 6, 1 << 7}

// BitMap implement a bitmap
type BitMap struct {
	Bits     []byte `json:"bits"`
	Count    uint   `json:"count"`
	Capacity uint   `json:"cap"`
}

// NewBitMap Create a new BitMap
func NewBitMap(cap uint) *BitMap {
	bits := make([]byte, (cap>>3)+1)
	return &BitMap{Bits: bits, Capacity: cap, Count: 0}
}

func (b *BitMap) Set(num uint) {
	byteIndex, bitIndex := b.offset(num)
	// move 1 left to the specified position and do the `|` operation
	b.Bits[byteIndex] |= 1 << bitIndex
	b.Count++
}

func (b *BitMap) Has(num uint) bool {
	byteIndex, bitIndex := b.offset(num)
	// 11110011 & 00000100 = 00000000
	return b.Bits[byteIndex]&(1<<bitIndex) != 0
}

func (b *BitMap) Reset(num uint) {
	byteIndex, bitIndex := b.offset(num)
	// find the position of the num and invert it
	// ^00000100 = 11111011
	// 11110011 & 11111011 = 11110011
	b.Bits[byteIndex] = b.Bits[byteIndex] & ^(1 << bitIndex)
	b.Count--
}

func (b *BitMap) IsFull() bool {
	return b.Count == b.Capacity
}

// First returns the index of the first zero value
func (b *BitMap) First() int {
	for i := uint(0); i < b.Capacity; i++ {
		if !b.Has(i) {
			return int(i)
		}
	}
	return -1
}

func (b *BitMap) String() string {
	var buffer strings.Builder
	for index := len(b.Bits) - 1; index >= 0; index-- {
		buffer.WriteString(byteToString(b.Bits[index]))
		buffer.WriteString(" ")
	}
	return buffer.String()
}

func (b *BitMap) offset(num uint) (byteIndex, bitIndex uint) {
	// byteIndex := num / 8
	byteIndex = num >> 3
	if byteIndex >= uint(len(b.Bits)) {
		panic(fmt.Sprintf("index value %d out of range", num))
	}
	// bitIndex := num % 8
	bitIndex = num & 0x07
	return
}

func byteToString(data byte) string {
	var buffer strings.Builder
	for index := 0; index < 8; index++ {
		if (bitmask[7-index] & data) == 0 {
			buffer.WriteString("0")
		} else {
			buffer.WriteString("1")
		}
	}
	return buffer.String()
}
