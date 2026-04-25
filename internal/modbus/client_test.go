package modbus

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

func TestServiceReadWriteAndFrameLogs(t *testing.T) {
	address, closeServer := startFakeServer(t)
	defer closeServer()

	service := NewService()
	status, err := service.Connect(ConnectRequest{
		Host:      "127.0.0.1",
		Port:      address.Port,
		UnitID:    1,
		TimeoutMs: 1000,
	})
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if !status.Connected {
		t.Fatalf("Connected = false, want true")
	}

	response, err := service.ReadRegisters(ReadRegistersRequest{
		FunctionCode: FunctionReadHoldingRegisters,
		Address:      0,
		Quantity:     2,
		DataType:     DataTypeFloat32,
		ByteOrder:    ByteOrderABCD,
	})
	if err != nil {
		t.Fatalf("ReadRegisters() error = %v", err)
	}
	if len(response.Rows) != 1 {
		t.Fatalf("len(response.Rows) = %d, want 1", len(response.Rows))
	}
	if response.Rows[0].Float32 == nil || *response.Rows[0].Float32 != 123 {
		t.Fatalf("Float32 = %v, want 123", response.Rows[0].Float32)
	}

	bits, err := service.ReadBits(ReadBitsRequest{
		FunctionCode: FunctionReadCoils,
		Address:      0,
		Quantity:     10,
		AddressBase:  1,
	})
	if err != nil {
		t.Fatalf("ReadBits(FC01) error = %v", err)
	}
	if len(bits.Rows) != 10 {
		t.Fatalf("len(bits.Rows) = %d, want 10", len(bits.Rows))
	}
	if bits.Rows[0].DisplayAddress != 1 || !bits.Rows[0].Value || bits.Rows[1].Value || !bits.Rows[8].Value {
		t.Fatalf("unexpected FC01 bit rows: first=%+v second=%+v ninth=%+v", bits.Rows[0], bits.Rows[1], bits.Rows[8])
	}

	inputs, err := service.ReadBits(ReadBitsRequest{
		FunctionCode: FunctionReadDiscreteInputs,
		Address:      0,
		Quantity:     8,
	})
	if err != nil {
		t.Fatalf("ReadBits(FC02) error = %v", err)
	}
	if len(inputs.Rows) != 8 {
		t.Fatalf("len(inputs.Rows) = %d, want 8", len(inputs.Rows))
	}
	if inputs.Rows[0].Value || !inputs.Rows[1].Value {
		t.Fatalf("unexpected FC02 bit rows: first=%+v second=%+v", inputs.Rows[0], inputs.Rows[1])
	}

	preview, err := service.PreviewWriteSingleRegister(WriteSingleRegisterRequest{Address: 10, Value: 230})
	if err != nil {
		t.Fatalf("PreviewWriteSingleRegister() error = %v", err)
	}
	if preview.RegisterHex != "0x00E6" {
		t.Fatalf("RegisterHex = %q, want 0x00E6", preview.RegisterHex)
	}
	if preview.Host != "127.0.0.1" || preview.Port != address.Port || preview.UnitID != 1 {
		t.Fatalf("preview target = %s:%d unit %d, want 127.0.0.1:%d unit 1", preview.Host, preview.Port, preview.UnitID, address.Port)
	}

	if err := service.WriteSingleRegister(WriteSingleRegisterRequest{Address: 10, Value: 230}); err != nil {
		t.Fatalf("WriteSingleRegister() error = %v", err)
	}

	coilPreview, err := service.PreviewWriteSingleCoil(WriteSingleCoilRequest{Address: 1, Value: true})
	if err != nil {
		t.Fatalf("PreviewWriteSingleCoil() error = %v", err)
	}
	if coilPreview.RegisterHex != "0xFF00" || coilPreview.Values[0] != "ON" {
		t.Fatalf("coil preview = %+v, want ON 0xFF00", coilPreview)
	}
	if err := service.WriteSingleCoil(WriteSingleCoilRequest{Address: 1, Value: true}); err != nil {
		t.Fatalf("WriteSingleCoil() error = %v", err)
	}

	coilsPreview, err := service.PreviewWriteMultipleCoils(WriteMultipleCoilsRequest{
		Address: 0,
		Values:  []bool{true, false, true, true, false, false, false, true, true},
	})
	if err != nil {
		t.Fatalf("PreviewWriteMultipleCoils() error = %v", err)
	}
	if coilsPreview.PDUHex != "0F 00 00 00 09 02 8D 01" {
		t.Fatalf("multiple coils PDUHex = %q, want packed bits", coilsPreview.PDUHex)
	}
	if err := service.WriteMultipleCoils(WriteMultipleCoilsRequest{
		Address: 0,
		Values:  []bool{true, false, true, true, false, false, false, true, true},
	}); err != nil {
		t.Fatalf("WriteMultipleCoils() error = %v", err)
	}

	registersPreview, err := service.PreviewWriteMultipleRegisters(WriteMultipleRegistersRequest{
		Address: 20,
		Values:  []uint16{0x1234, 0xABCD},
	})
	if err != nil {
		t.Fatalf("PreviewWriteMultipleRegisters() error = %v", err)
	}
	if registersPreview.RegisterHex != "0x1234, 0xABCD" {
		t.Fatalf("RegisterHex = %q, want 0x1234, 0xABCD", registersPreview.RegisterHex)
	}
	if err := service.WriteMultipleRegisters(WriteMultipleRegistersRequest{
		Address: 20,
		Values:  []uint16{0x1234, 0xABCD},
	}); err != nil {
		t.Fatalf("WriteMultipleRegisters() error = %v", err)
	}

	logs := service.FrameLogs()
	if len(logs) < 14 {
		t.Fatalf("len(logs) = %d, want at least 14", len(logs))
	}
	if logs[0].Direction != "TX" || logs[1].Direction != "RX" {
		t.Fatalf("first log directions = %s/%s, want TX/RX", logs[0].Direction, logs[1].Direction)
	}
}

func TestPreviewWriteSingleRegisterRequiresConnection(t *testing.T) {
	service := NewService()
	_, err := service.PreviewWriteSingleRegister(WriteSingleRegisterRequest{Address: 10, Value: 230})
	if err == nil {
		t.Fatal("PreviewWriteSingleRegister() error = nil, want error")
	}
}

func TestModbusExceptionIsNotRetried(t *testing.T) {
	var requests atomic.Int32
	address, closeServer := startScriptedServer(t, func(header []byte, body []byte) ([]byte, bool) {
		requests.Add(1)
		return []byte{body[0] | 0x80, 0x02}, true
	})
	defer closeServer()

	service := NewService()
	_, err := service.Connect(ConnectRequest{
		Host:      "127.0.0.1",
		Port:      address.Port,
		UnitID:    1,
		TimeoutMs: 1000,
		Retries:   3,
	})
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	_, err = service.ReadRegisters(ReadRegistersRequest{
		FunctionCode: FunctionReadHoldingRegisters,
		Address:      0,
		Quantity:     2,
		DataType:     DataTypeUInt16,
		ByteOrder:    ByteOrderABCD,
	})
	var exceptionErr *ExceptionError
	if !errors.As(err, &exceptionErr) {
		t.Fatalf("ReadRegisters() error = %T %v, want ExceptionError", err, err)
	}
	if got := requests.Load(); got != 1 {
		t.Fatalf("requests = %d, want 1", got)
	}
	if !service.Status().Connected {
		t.Fatal("Connected = false, want true after Modbus exception")
	}
}

func TestWritesDoNotRetryTransportFailure(t *testing.T) {
	tests := []struct {
		name  string
		write func(*Service) error
	}{
		{
			name: "FC05",
			write: func(service *Service) error {
				return service.WriteSingleCoil(WriteSingleCoilRequest{Address: 1, Value: true})
			},
		},
		{
			name: "FC06",
			write: func(service *Service) error {
				return service.WriteSingleRegister(WriteSingleRegisterRequest{Address: 10, Value: 230})
			},
		},
		{
			name: "FC15",
			write: func(service *Service) error {
				return service.WriteMultipleCoils(WriteMultipleCoilsRequest{Address: 0, Values: []bool{true, false}})
			},
		},
		{
			name: "FC16",
			write: func(service *Service) error {
				return service.WriteMultipleRegisters(WriteMultipleRegistersRequest{Address: 20, Values: []uint16{1, 2}})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requests atomic.Int32
			address, closeServer := startScriptedServer(t, func(header []byte, body []byte) ([]byte, bool) {
				requests.Add(1)
				return nil, false
			})
			defer closeServer()

			service := NewService()
			_, err := service.Connect(ConnectRequest{
				Host:      "127.0.0.1",
				Port:      address.Port,
				UnitID:    1,
				TimeoutMs: 200,
				Retries:   3,
			})
			if err != nil {
				t.Fatalf("Connect() error = %v", err)
			}

			err = tt.write(service)
			if err == nil {
				t.Fatal("write error = nil, want error")
			}
			if got := requests.Load(); got != 1 {
				t.Fatalf("requests = %d, want 1", got)
			}
		})
	}
}

type fakeServerAddress struct {
	Port int
}

func startFakeServer(t *testing.T) (fakeServerAddress, func()) {
	t.Helper()
	return startScriptedServer(t, func(header []byte, body []byte) ([]byte, bool) {
		functionCode := body[0]
		switch functionCode {
		case FunctionReadCoils:
			return []byte{FunctionReadCoils, 0x02, 0x8D, 0x01}, true
		case FunctionReadDiscreteInputs:
			return []byte{FunctionReadDiscreteInputs, 0x01, 0x72}, true
		case FunctionReadHoldingRegisters:
			return []byte{FunctionReadHoldingRegisters, 0x04, 0x42, 0xF6, 0x00, 0x00}, true
		case FunctionWriteSingleCoil:
			return append([]byte{FunctionWriteSingleCoil}, body[1:5]...), true
		case FunctionWriteSingleRegister:
			return append([]byte{FunctionWriteSingleRegister}, body[1:5]...), true
		case FunctionWriteMultipleCoils:
			return append([]byte{FunctionWriteMultipleCoils}, body[1:5]...), true
		case FunctionWriteMultipleRegisters:
			return append([]byte{FunctionWriteMultipleRegisters}, body[1:5]...), true
		default:
			return []byte{functionCode | 0x80, 0x01}, true
		}
	})
}

func startScriptedServer(t *testing.T, handler func(header []byte, body []byte) ([]byte, bool)) (fakeServerAddress, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

		for {
			header := make([]byte, 7)
			if _, err := io.ReadFull(conn, header); err != nil {
				return
			}
			length := int(binary.BigEndian.Uint16(header[4:6]))
			body := make([]byte, length-1)
			if _, err := io.ReadFull(conn, body); err != nil {
				return
			}

			txID := binary.BigEndian.Uint16(header[0:2])
			unitID := header[6]
			pdu, reply := handler(header, body)
			if !reply {
				return
			}

			if _, err := conn.Write(buildFrame(txID, unitID, pdu)); err != nil {
				return
			}
		}
	}()

	tcpAddr := listener.Addr().(*net.TCPAddr)
	closeFn := func() {
		_ = listener.Close()
		select {
		case <-done:
		case <-time.After(time.Second):
		}
	}
	return fakeServerAddress{Port: tcpAddr.Port}, closeFn
}
