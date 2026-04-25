package modbus

import (
	"math"
	"testing"
	"time"
)

func TestDecodeFloat32(t *testing.T) {
	tests := []struct {
		name      string
		registers []uint16
		order     string
		want      float32
	}{
		{name: "ABCD", registers: []uint16{0x42F6, 0x0000}, order: ByteOrderABCD, want: 123},
		{name: "CDAB", registers: []uint16{0x0000, 0x42F6}, order: ByteOrderCDAB, want: 123},
		{name: "BADC", registers: []uint16{0xF642, 0x0000}, order: ByteOrderBADC, want: 123},
		{name: "DCBA", registers: []uint16{0x0000, 0xF642}, order: ByteOrderDCBA, want: 123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeFloat32(tt.registers, tt.order)
			if err != nil {
				t.Fatalf("DecodeFloat32() error = %v", err)
			}
			if math.Abs(float64(got-tt.want)) > 0.0001 {
				t.Fatalf("DecodeFloat32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeRegisterRows(t *testing.T) {
	now := time.Date(2026, 4, 24, 12, 31, 2, 123000000, time.UTC)
	rows, err := DecodeRegisterRows(10, []uint16{0xFFFF, 0x42F6, 0x0000}, DataTypeInt16, ByteOrderABCD, 1, now)
	if err != nil {
		t.Fatalf("DecodeRegisterRows() error = %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("len(rows) = %d, want 3", len(rows))
	}
	if rows[0].DisplayAddress != 11 {
		t.Fatalf("DisplayAddress = %d, want 11", rows[0].DisplayAddress)
	}
	if rows[0].Int16 != -1 {
		t.Fatalf("Int16 = %d, want -1", rows[0].Int16)
	}
	if rows[0].RawHex != "0xFFFF" {
		t.Fatalf("RawHex = %q, want 0xFFFF", rows[0].RawHex)
	}
	if rows[0].Timestamp != "12:31:02.123" {
		t.Fatalf("Timestamp = %q, want 12:31:02.123", rows[0].Timestamp)
	}
}

func TestDecodeRegisterRowsFloat32Incomplete(t *testing.T) {
	rows, err := DecodeRegisterRows(0, []uint16{0x42F6}, DataTypeFloat32, ByteOrderABCD, 0, time.Now())
	if err != nil {
		t.Fatalf("DecodeRegisterRows() error = %v", err)
	}
	if rows[0].Quality != "Incomplete" {
		t.Fatalf("Quality = %q, want Incomplete", rows[0].Quality)
	}
}

func TestDecodeRegisterRowsFloat32NonOverlapping(t *testing.T) {
	rows, err := DecodeRegisterRows(3000, []uint16{0x42F6, 0x0000, 0x4120, 0x0000}, DataTypeFloat32, ByteOrderABCD, 1, time.Now())
	if err != nil {
		t.Fatalf("DecodeRegisterRows() error = %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2", len(rows))
	}
	if rows[0].DisplayAddress != 3001 || rows[0].RegisterCount != 2 {
		t.Fatalf("first row address/count = %d/%d, want 3001/2", rows[0].DisplayAddress, rows[0].RegisterCount)
	}
	if rows[1].DisplayAddress != 3003 || rows[1].RegisterCount != 2 {
		t.Fatalf("second row address/count = %d/%d, want 3003/2", rows[1].DisplayAddress, rows[1].RegisterCount)
	}
	if rows[0].Float32 == nil || *rows[0].Float32 != 123 {
		t.Fatalf("first Float32 = %v, want 123", rows[0].Float32)
	}
	if rows[1].Float32 == nil || *rows[1].Float32 != 10 {
		t.Fatalf("second Float32 = %v, want 10", rows[1].Float32)
	}
}

func TestUnpackBits(t *testing.T) {
	got := UnpackBits([]byte{0x8D, 0x01}, 10)
	want := []bool{true, false, true, true, false, false, false, true, true, false}
	if len(got) != len(want) {
		t.Fatalf("len(UnpackBits()) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("bit %d = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestDecodeBitRows(t *testing.T) {
	now := time.Date(2026, 4, 25, 7, 45, 1, 456000000, time.UTC)
	rows := DecodeBitRows(20, []bool{true, false}, 1, now)
	if len(rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2", len(rows))
	}
	if rows[0].DisplayAddress != 21 || !rows[0].Value || rows[0].Raw != "1" {
		t.Fatalf("first row = %+v, want display 21 true raw 1", rows[0])
	}
	if rows[1].DisplayAddress != 22 || rows[1].Value || rows[1].Raw != "0" {
		t.Fatalf("second row = %+v, want display 22 false raw 0", rows[1])
	}
	if rows[0].Timestamp != "07:45:01.456" {
		t.Fatalf("Timestamp = %q, want 07:45:01.456", rows[0].Timestamp)
	}
}
