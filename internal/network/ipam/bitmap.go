package ipam

import (
	"fmt"
	"strings"
)

// [1 2 4 8 16 32 64 128]
var bitmask = []byte{1, 1 << 1, 1 << 2, 1 << 3, 1 << 4, 1 << 5, 1 << 6, 1 << 7}

// BitMap implement a bitmap
type BitMap struct {
	bits     []byte
	count    uint
	capacity uint
}

// NewBitMap Create a new BitMap
func NewBitMap(cap int) *BitMap {
	bits := make([]byte, (cap>>3)+1)
	return &BitMap{bits: bits, capacity: uint(cap), count: 0}
}

func (b *BitMap) Set(num uint) {
	byteIndex, bitIndex := b.offset(num)
	// move 1 left to the specified position and do the `|` operation
	b.bits[byteIndex] |= 1 << bitIndex
	b.count++
}

func (b *BitMap) Has(num uint) bool {
	byteIndex, bitIndex := b.offset(num)
	// 11110011 & 00000100 = 00000000
	return b.bits[byteIndex]&(1<<bitIndex) != 0
}

func (b *BitMap) Reset(num uint) {
	byteIndex, bitIndex := b.offset(num)
	// find the position of the num and invert it
	// ^00000100 = 11111011
	// 11110011 & 11111011 = 11110011
	b.bits[byteIndex] = b.bits[byteIndex] & ^(1 << bitIndex)
	b.count--
}

func (b *BitMap) Count() int {
	return int(b.count)
}

func (b *BitMap) IsFull() bool {
	return b.count == b.capacity
}

// SetFirst returns the index of the first zero value
func (b *BitMap) SetFirst() {
	for byteIndex := len(b.bits) - 1; byteIndex >= 0; byteIndex-- {
		for bitIndex := 0; bitIndex < 8; bitIndex++ {
			if (bitmask[7-bitIndex] & b.bits[byteIndex]) == 0 {
				b.bits[byteIndex] |= 1 << bitIndex
				return
			}
		}
	}
}

func (b *BitMap) String() string {
	var buffer strings.Builder
	for index := len(b.bits) - 1; index >= 0; index-- {
		buffer.WriteString(byteToString(b.bits[index]))
		buffer.WriteString(" ")
	}
	return buffer.String()
}

func (b *BitMap) offset(num uint) (byteIndex, bitIndex uint) {
	// byteIndex := num / 8
	byteIndex = num >> 3
	if byteIndex >= uint(len(b.bits)) {
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
