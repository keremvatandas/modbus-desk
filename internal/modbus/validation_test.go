package modbus

import "testing"

func TestValidateConnectRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     ConnectRequest
		wantErr bool
	}{
		{name: "valid default port", req: ConnectRequest{Host: "192.168.1.10", UnitID: 1, TimeoutMs: 1000}, wantErr: false},
		{name: "missing host", req: ConnectRequest{UnitID: 1, TimeoutMs: 1000}, wantErr: true},
		{name: "invalid port", req: ConnectRequest{Host: "localhost", Port: 70000, UnitID: 1, TimeoutMs: 1000}, wantErr: true},
		{name: "invalid unit id", req: ConnectRequest{Host: "localhost", Port: 502, UnitID: 248, TimeoutMs: 1000}, wantErr: true},
		{name: "invalid timeout", req: ConnectRequest{Host: "localhost", Port: 502, UnitID: 1, TimeoutMs: 99}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConnectRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateConnectRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateReadRegistersRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     ReadRegistersRequest
		wantErr bool
	}{
		{name: "valid FC03", req: ReadRegistersRequest{FunctionCode: 3, Quantity: 2, DataType: "uint16", ByteOrder: "ABCD"}, wantErr: false},
		{name: "valid FC04", req: ReadRegistersRequest{FunctionCode: 4, Quantity: 2, DataType: "float32", ByteOrder: "CDAB"}, wantErr: false},
		{name: "invalid function", req: ReadRegistersRequest{FunctionCode: 1, Quantity: 2, DataType: "uint16", ByteOrder: "ABCD"}, wantErr: true},
		{name: "zero quantity", req: ReadRegistersRequest{FunctionCode: 3, Quantity: 0, DataType: "uint16", ByteOrder: "ABCD"}, wantErr: true},
		{name: "too many registers", req: ReadRegistersRequest{FunctionCode: 3, Quantity: 126, DataType: "uint16", ByteOrder: "ABCD"}, wantErr: true},
		{name: "address range overflow", req: ReadRegistersRequest{FunctionCode: 3, Address: 65535, Quantity: 2, DataType: "uint16", ByteOrder: "ABCD"}, wantErr: true},
		{name: "bad type", req: ReadRegistersRequest{FunctionCode: 3, Quantity: 2, DataType: "float64", ByteOrder: "ABCD"}, wantErr: true},
		{name: "bad order", req: ReadRegistersRequest{FunctionCode: 3, Quantity: 2, DataType: "uint16", ByteOrder: "ACBD"}, wantErr: true},
		{name: "bad address base", req: ReadRegistersRequest{FunctionCode: 3, Quantity: 2, DataType: "uint16", ByteOrder: "ABCD", AddressBase: 2}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateReadRegistersRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateReadRegistersRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateReadBitsRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     ReadBitsRequest
		wantErr bool
	}{
		{name: "valid FC01", req: ReadBitsRequest{FunctionCode: FunctionReadCoils, Quantity: 10}, wantErr: false},
		{name: "valid FC02", req: ReadBitsRequest{FunctionCode: FunctionReadDiscreteInputs, Quantity: 2000, AddressBase: 1}, wantErr: false},
		{name: "invalid function", req: ReadBitsRequest{FunctionCode: FunctionReadHoldingRegisters, Quantity: 10}, wantErr: true},
		{name: "zero quantity", req: ReadBitsRequest{FunctionCode: FunctionReadCoils, Quantity: 0}, wantErr: true},
		{name: "too many bits", req: ReadBitsRequest{FunctionCode: FunctionReadCoils, Quantity: 2001}, wantErr: true},
		{name: "address range overflow", req: ReadBitsRequest{FunctionCode: FunctionReadCoils, Address: 65535, Quantity: 2}, wantErr: true},
		{name: "bad address base", req: ReadBitsRequest{FunctionCode: FunctionReadCoils, Quantity: 10, AddressBase: 2}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateReadBitsRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateReadBitsRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateWriteSingleCoilRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     WriteSingleCoilRequest
		wantErr bool
	}{
		{name: "valid", req: WriteSingleCoilRequest{Address: 65535, Value: true}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWriteSingleCoilRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateWriteSingleCoilRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateWriteMultipleCoilsRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     WriteMultipleCoilsRequest
		wantErr bool
	}{
		{name: "valid", req: WriteMultipleCoilsRequest{Address: 10, Values: []bool{true, false, true}}, wantErr: false},
		{name: "empty values", req: WriteMultipleCoilsRequest{Address: 10}, wantErr: true},
		{name: "too many coils", req: WriteMultipleCoilsRequest{Values: make([]bool, 1969)}, wantErr: true},
		{name: "address range overflow", req: WriteMultipleCoilsRequest{Address: 65535, Values: []bool{true, false}}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWriteMultipleCoilsRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateWriteMultipleCoilsRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateWriteMultipleRegistersRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     WriteMultipleRegistersRequest
		wantErr bool
	}{
		{name: "valid", req: WriteMultipleRegistersRequest{Address: 10, Values: []uint16{1, 2, 3}}, wantErr: false},
		{name: "empty values", req: WriteMultipleRegistersRequest{Address: 10}, wantErr: true},
		{name: "too many registers", req: WriteMultipleRegistersRequest{Values: make([]uint16, 124)}, wantErr: true},
		{name: "address range overflow", req: WriteMultipleRegistersRequest{Address: 65535, Values: []uint16{1, 2}}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWriteMultipleRegistersRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateWriteMultipleRegistersRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
