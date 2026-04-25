package modbus

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"
)

func DecodeRegisterRows(address uint16, registers []uint16, dataType string, byteOrder string, addressBase int, now time.Time) ([]RegisterRow, error) {
	dataType = normalizeDataType(dataType)
	byteOrder = normalizeByteOrder(byteOrder)
	if err := ValidateDataType(dataType); err != nil {
		return nil, err
	}
	if err := ValidateByteOrder(byteOrder); err != nil {
		return nil, err
	}

	if dataType == DataTypeFloat32 {
		return decodeFloat32Rows(address, registers, byteOrder, addressBase, now)
	}

	rows := make([]RegisterRow, 0, len(registers))
	for i, reg := range registers {
		rowAddress := address + uint16(i)
		row := RegisterRow{
			Address:        rowAddress,
			DisplayAddress: int(rowAddress) + addressBase,
			RegisterCount:  1,
			RawHex:         fmt.Sprintf("0x%04X", reg),
			UInt16:         reg,
			Int16:          int16(reg),
			Binary:         fmt.Sprintf("%016b", reg),
			DataType:       dataType,
			Timestamp:      timestamp(now),
			Quality:        "OK",
		}

		switch dataType {
		case DataTypeUInt16:
			row.Value = reg
		case DataTypeInt16:
			row.Value = int16(reg)
		case DataTypeHex:
			row.Value = row.RawHex
		case DataTypeBinary:
			row.Value = row.Binary
		}

		rows = append(rows, row)
	}
	return rows, nil
}

func DecodeBitRows(address uint16, bits []bool, addressBase int, now time.Time) []BitRow {
	rows := make([]BitRow, 0, len(bits))
	for i, bit := range bits {
		rowAddress := address + uint16(i)
		raw := "0"
		if bit {
			raw = "1"
		}
		rows = append(rows, BitRow{
			Address:        rowAddress,
			DisplayAddress: int(rowAddress) + addressBase,
			Raw:            raw,
			Value:          bit,
			Timestamp:      timestamp(now),
			Quality:        "OK",
		})
	}
	return rows
}

func UnpackBits(data []byte, quantity uint16) []bool {
	bits := make([]bool, 0, quantity)
	for _, b := range data {
		for bit := 0; bit < 8 && len(bits) < int(quantity); bit++ {
			bits = append(bits, b&(1<<bit) != 0)
		}
	}
	return bits
}

func decodeFloat32Rows(address uint16, registers []uint16, byteOrder string, addressBase int, now time.Time) ([]RegisterRow, error) {
	rows := make([]RegisterRow, 0, (len(registers)+1)/2)
	for i := 0; i < len(registers); i += 2 {
		rowAddress := address + uint16(i)
		row := RegisterRow{
			Address:        rowAddress,
			DisplayAddress: int(rowAddress) + addressBase,
			RegisterCount:  2,
			UInt16:         registers[i],
			Int16:          int16(registers[i]),
			Binary:         fmt.Sprintf("%016b", registers[i]),
			DataType:       DataTypeFloat32,
			Timestamp:      timestamp(now),
			Quality:        "OK",
		}
		if i+1 >= len(registers) {
			row.RegisterCount = 1
			row.RawHex = fmt.Sprintf("0x%04X", registers[i])
			row.Quality = "Incomplete"
			row.Value = nil
			rows = append(rows, row)
			continue
		}

		pair := registers[i : i+2]
		value, err := DecodeFloat32(pair, byteOrder)
		if err != nil {
			return nil, err
		}
		float64Value := float64(value)
		row.RawHex = "0x" + strings.ReplaceAll(registersToHex(pair), " ", " 0x")
		row.Binary = fmt.Sprintf("%016b %016b", pair[0], pair[1])
		row.Float32 = &float64Value
		row.Value = float64Value
		rows = append(rows, row)
	}
	return rows, nil
}

func DecodeFloat32(registers []uint16, byteOrder string) (float32, error) {
	if len(registers) < 2 {
		return 0, fmt.Errorf("float32 requires at least 2 registers")
	}
	ordered, err := orderedBytes(registers[:2], byteOrder)
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.BigEndian.Uint32(ordered)), nil
}

func orderedBytes(registers []uint16, byteOrder string) ([]byte, error) {
	byteOrder = normalizeByteOrder(byteOrder)
	if len(registers) < 2 {
		return nil, fmt.Errorf("at least 2 registers are required")
	}

	bytes := []byte{
		byte(registers[0] >> 8),
		byte(registers[0]),
		byte(registers[1] >> 8),
		byte(registers[1]),
	}

	switch byteOrder {
	case ByteOrderABCD:
		return bytes, nil
	case ByteOrderBADC:
		return []byte{bytes[1], bytes[0], bytes[3], bytes[2]}, nil
	case ByteOrderCDAB:
		return []byte{bytes[2], bytes[3], bytes[0], bytes[1]}, nil
	case ByteOrderDCBA:
		return []byte{bytes[3], bytes[2], bytes[1], bytes[0]}, nil
	default:
		return nil, fmt.Errorf("unsupported byte order %q", byteOrder)
	}
}

func registersToHex(registers []uint16) string {
	parts := make([]string, 0, len(registers))
	for _, register := range registers {
		parts = append(parts, fmt.Sprintf("%04X", register))
	}
	return strings.Join(parts, " ")
}
