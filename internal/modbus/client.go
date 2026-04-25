package modbus

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Service struct {
	mu             sync.Mutex
	conn           net.Conn
	config         ConnectRequest
	transactionID  uint16
	logs           []FrameLog
	nextLogID      int
	lastError      string
	lastResponseMs float64
}

type writePreviewRequest struct {
	FunctionCode int
	Address      uint16
	Quantity     uint16
	Value        uint16
	Values       []string
	RegisterHex  string
	pdu          []byte
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Connect(req ConnectRequest) (ConnectionStatus, error) {
	req = NormalizeConnectRequest(req)
	if err := ValidateConnectRequest(req); err != nil {
		s.setLastError(err)
		return s.Status(), err
	}

	conn, err := dial(req)
	if err != nil {
		wrapped := fmt.Errorf("connect to %s:%d failed: %w", req.Host, req.Port, err)
		s.setLastError(wrapped)
		return s.Status(), wrapped
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != nil {
		_ = s.conn.Close()
	}
	s.conn = conn
	s.config = req
	s.lastError = ""
	s.lastResponseMs = 0
	return s.statusLocked(), nil
}

func (s *Service) Disconnect() ConnectionStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != nil {
		_ = s.conn.Close()
	}
	s.conn = nil
	s.lastError = ""
	return s.statusLocked()
}

func (s *Service) Status() ConnectionStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.statusLocked()
}

func (s *Service) ReadRegisters(req ReadRegistersRequest) (ReadRegistersResponse, error) {
	start := time.Now()
	if err := ValidateReadRegistersRequest(req); err != nil {
		s.setLastError(err)
		return ReadRegistersResponse{}, err
	}

	payload := make([]byte, 4)
	binary.BigEndian.PutUint16(payload[0:2], req.Address)
	binary.BigEndian.PutUint16(payload[2:4], req.Quantity)

	response, err := s.send(byte(req.FunctionCode), payload, true)
	if err != nil {
		s.setLastError(err)
		return ReadRegistersResponse{}, err
	}
	if len(response) < 1 {
		err := fmt.Errorf("empty register response")
		s.setLastError(err)
		return ReadRegistersResponse{}, err
	}

	byteCount := int(response[0])
	if byteCount != int(req.Quantity)*2 {
		err := fmt.Errorf("unexpected register byte count %d, expected %d", byteCount, req.Quantity*2)
		s.setLastError(err)
		return ReadRegistersResponse{}, err
	}
	if len(response[1:]) < byteCount {
		err := fmt.Errorf("short register response")
		s.setLastError(err)
		return ReadRegistersResponse{}, err
	}

	registers := make([]uint16, req.Quantity)
	for i := 0; i < int(req.Quantity); i++ {
		offset := 1 + i*2
		registers[i] = binary.BigEndian.Uint16(response[offset : offset+2])
	}

	rows, err := DecodeRegisterRows(req.Address, registers, req.DataType, req.ByteOrder, req.AddressBase, time.Now())
	if err != nil {
		s.setLastError(err)
		return ReadRegistersResponse{}, err
	}

	duration := float64(time.Since(start).Microseconds()) / 1000
	s.mu.Lock()
	s.lastError = ""
	s.lastResponseMs = duration
	s.mu.Unlock()

	return ReadRegistersResponse{
		FunctionCode: req.FunctionCode,
		Address:      req.Address,
		Quantity:     req.Quantity,
		Rows:         rows,
		DurationMs:   duration,
	}, nil
}

func (s *Service) ReadBits(req ReadBitsRequest) (ReadBitsResponse, error) {
	start := time.Now()
	if err := ValidateReadBitsRequest(req); err != nil {
		s.setLastError(err)
		return ReadBitsResponse{}, err
	}

	payload := make([]byte, 4)
	binary.BigEndian.PutUint16(payload[0:2], req.Address)
	binary.BigEndian.PutUint16(payload[2:4], req.Quantity)

	response, err := s.send(byte(req.FunctionCode), payload, true)
	if err != nil {
		s.setLastError(err)
		return ReadBitsResponse{}, err
	}
	if len(response) < 1 {
		err := fmt.Errorf("empty bit response")
		s.setLastError(err)
		return ReadBitsResponse{}, err
	}

	byteCount := int(response[0])
	expectedByteCount := int((req.Quantity + 7) / 8)
	if byteCount != expectedByteCount {
		err := fmt.Errorf("unexpected bit byte count %d, expected %d", byteCount, expectedByteCount)
		s.setLastError(err)
		return ReadBitsResponse{}, err
	}
	if len(response[1:]) < byteCount {
		err := fmt.Errorf("short bit response")
		s.setLastError(err)
		return ReadBitsResponse{}, err
	}

	bits := UnpackBits(response[1:1+byteCount], req.Quantity)
	rows := DecodeBitRows(req.Address, bits, req.AddressBase, time.Now())

	duration := float64(time.Since(start).Microseconds()) / 1000
	s.mu.Lock()
	s.lastError = ""
	s.lastResponseMs = duration
	s.mu.Unlock()

	return ReadBitsResponse{
		FunctionCode: req.FunctionCode,
		Address:      req.Address,
		Quantity:     req.Quantity,
		Rows:         rows,
		DurationMs:   duration,
	}, nil
}

func (s *Service) PreviewWriteSingleRegister(req WriteSingleRegisterRequest) (WritePreview, error) {
	if err := ValidateWriteSingleRegisterRequest(req); err != nil {
		return WritePreview{}, err
	}

	pdu := writeSingleRegisterPDU(req)
	return s.previewWrite(writePreviewRequest{
		FunctionCode: FunctionWriteSingleRegister,
		Address:      req.Address,
		Quantity:     1,
		Value:        req.Value,
		RegisterHex:  fmt.Sprintf("0x%04X", req.Value),
		Values:       []string{fmt.Sprintf("%d", req.Value)},
		pdu:          pdu,
	})
}

func (s *Service) PreviewWriteSingleCoil(req WriteSingleCoilRequest) (WritePreview, error) {
	if err := ValidateWriteSingleCoilRequest(req); err != nil {
		return WritePreview{}, err
	}

	pdu := writeSingleCoilPDU(req)
	return s.previewWrite(writePreviewRequest{
		FunctionCode: FunctionWriteSingleCoil,
		Address:      req.Address,
		Quantity:     1,
		Value:        coilPreviewValue(req.Value),
		RegisterHex:  fmt.Sprintf("0x%04X", coilWireValue(req.Value)),
		Values:       []string{coilLabel(req.Value)},
		pdu:          pdu,
	})
}

func (s *Service) PreviewWriteMultipleCoils(req WriteMultipleCoilsRequest) (WritePreview, error) {
	if err := ValidateWriteMultipleCoilsRequest(req); err != nil {
		return WritePreview{}, err
	}

	pdu := writeMultipleCoilsPDU(req)
	return s.previewWrite(writePreviewRequest{
		FunctionCode: FunctionWriteMultipleCoils,
		Address:      req.Address,
		Quantity:     uint16(len(req.Values)),
		Values:       coilLabels(req.Values),
		pdu:          pdu,
	})
}

func (s *Service) PreviewWriteMultipleRegisters(req WriteMultipleRegistersRequest) (WritePreview, error) {
	if err := ValidateWriteMultipleRegistersRequest(req); err != nil {
		return WritePreview{}, err
	}

	pdu := writeMultipleRegistersPDU(req)
	return s.previewWrite(writePreviewRequest{
		FunctionCode: FunctionWriteMultipleRegisters,
		Address:      req.Address,
		Quantity:     uint16(len(req.Values)),
		RegisterHex:  registerValuesHex(req.Values),
		Values:       registerValueLabels(req.Values),
		pdu:          pdu,
	})
}

func (s *Service) previewWrite(req writePreviewRequest) (WritePreview, error) {
	s.mu.Lock()
	unitID := s.config.UnitID
	host := s.config.Host
	port := s.config.Port
	txID := s.transactionID + 1
	connected := s.conn != nil
	s.mu.Unlock()
	if !connected {
		return WritePreview{}, fmt.Errorf("not connected")
	}

	frame := buildFrame(txID, byte(unitID), req.pdu)
	return WritePreview{
		FunctionCode: req.FunctionCode,
		Address:      req.Address,
		Quantity:     req.Quantity,
		Value:        req.Value,
		Values:       req.Values,
		Host:         host,
		Port:         port,
		UnitID:       unitID,
		RegisterHex:  req.RegisterHex,
		PDUHex:       hexBytes(req.pdu),
		FrameHex:     hexBytes(frame),
	}, nil
}

func (s *Service) WriteSingleRegister(req WriteSingleRegisterRequest) error {
	if err := ValidateWriteSingleRegisterRequest(req); err != nil {
		s.setLastError(err)
		return err
	}

	request := writeSingleRegisterPDU(req)
	response, err := s.send(FunctionWriteSingleRegister, request[1:], false)
	if err != nil {
		s.setLastError(err)
		return err
	}
	if len(response) != 4 {
		err := fmt.Errorf("unexpected FC06 response length %d", len(response))
		s.setLastError(err)
		return err
	}
	if binary.BigEndian.Uint16(response[0:2]) != req.Address || binary.BigEndian.Uint16(response[2:4]) != req.Value {
		err := fmt.Errorf("FC06 response echo does not match request")
		s.setLastError(err)
		return err
	}

	s.mu.Lock()
	s.lastError = ""
	s.mu.Unlock()
	return nil
}

func (s *Service) WriteSingleCoil(req WriteSingleCoilRequest) error {
	if err := ValidateWriteSingleCoilRequest(req); err != nil {
		s.setLastError(err)
		return err
	}

	request := writeSingleCoilPDU(req)
	response, err := s.send(FunctionWriteSingleCoil, request[1:], false)
	if err != nil {
		s.setLastError(err)
		return err
	}
	if len(response) != 4 {
		err := fmt.Errorf("unexpected FC05 response length %d", len(response))
		s.setLastError(err)
		return err
	}
	if binary.BigEndian.Uint16(response[0:2]) != req.Address || binary.BigEndian.Uint16(response[2:4]) != coilWireValue(req.Value) {
		err := fmt.Errorf("FC05 response echo does not match request")
		s.setLastError(err)
		return err
	}

	s.mu.Lock()
	s.lastError = ""
	s.mu.Unlock()
	return nil
}

func (s *Service) WriteMultipleCoils(req WriteMultipleCoilsRequest) error {
	if err := ValidateWriteMultipleCoilsRequest(req); err != nil {
		s.setLastError(err)
		return err
	}

	request := writeMultipleCoilsPDU(req)
	response, err := s.send(FunctionWriteMultipleCoils, request[1:], false)
	if err != nil {
		s.setLastError(err)
		return err
	}
	if err := validateWriteQuantityEcho(FunctionWriteMultipleCoils, response, req.Address, uint16(len(req.Values))); err != nil {
		s.setLastError(err)
		return err
	}

	s.mu.Lock()
	s.lastError = ""
	s.mu.Unlock()
	return nil
}

func (s *Service) WriteMultipleRegisters(req WriteMultipleRegistersRequest) error {
	if err := ValidateWriteMultipleRegistersRequest(req); err != nil {
		s.setLastError(err)
		return err
	}

	request := writeMultipleRegistersPDU(req)
	response, err := s.send(FunctionWriteMultipleRegisters, request[1:], false)
	if err != nil {
		s.setLastError(err)
		return err
	}
	if err := validateWriteQuantityEcho(FunctionWriteMultipleRegisters, response, req.Address, uint16(len(req.Values))); err != nil {
		s.setLastError(err)
		return err
	}

	s.mu.Lock()
	s.lastError = ""
	s.mu.Unlock()
	return nil
}

func (s *Service) FrameLogs() []FrameLog {
	s.mu.Lock()
	defer s.mu.Unlock()
	logs := make([]FrameLog, len(s.logs))
	copy(logs, s.logs)
	return logs
}

func (s *Service) send(functionCode byte, payload []byte, allowRetry bool) ([]byte, error) {
	var lastErr error
	attempts := 1

	s.mu.Lock()
	if allowRetry && s.config.Retries > 0 {
		attempts += s.config.Retries
	}
	s.mu.Unlock()

	for attempt := 0; attempt < attempts; attempt++ {
		response, err := s.sendOnce(functionCode, payload)
		if err == nil {
			return response, nil
		}
		var exceptionErr *ExceptionError
		if errors.As(err, &exceptionErr) {
			return nil, err
		}
		lastErr = err
		s.closeConnection()
		if allowRetry && attempt+1 < attempts {
			if reconnectErr := s.reconnect(); reconnectErr != nil {
				lastErr = reconnectErr
				break
			}
		}
	}
	return nil, lastErr
}

func (s *Service) sendOnce(functionCode byte, payload []byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn == nil {
		return nil, fmt.Errorf("not connected")
	}
	if s.config.TimeoutMs <= 0 {
		s.config.TimeoutMs = 3000
	}

	s.transactionID++
	pdu := append([]byte{functionCode}, payload...)
	frame := buildFrame(s.transactionID, byte(s.config.UnitID), pdu)
	start := time.Now()
	s.appendFrameLogLocked("TX", frame, 0)

	deadline := time.Now().Add(time.Duration(s.config.TimeoutMs) * time.Millisecond)
	if err := s.conn.SetDeadline(deadline); err != nil {
		return nil, err
	}
	if _, err := s.conn.Write(frame); err != nil {
		return nil, err
	}

	header := make([]byte, 7)
	if _, err := io.ReadFull(s.conn, header); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint16(header[4:6])
	if length < 2 {
		return nil, fmt.Errorf("invalid MBAP length %d", length)
	}

	body := make([]byte, int(length)-1)
	if _, err := io.ReadFull(s.conn, body); err != nil {
		return nil, err
	}

	responseFrame := append(header, body...)
	duration := float64(time.Since(start).Microseconds()) / 1000
	s.lastResponseMs = duration
	s.appendFrameLogLocked("RX", responseFrame, duration)

	if txID := binary.BigEndian.Uint16(header[0:2]); txID != s.transactionID {
		return nil, fmt.Errorf("transaction ID mismatch: got %d, expected %d", txID, s.transactionID)
	}
	if protocolID := binary.BigEndian.Uint16(header[2:4]); protocolID != 0 {
		return nil, fmt.Errorf("unsupported protocol ID %d", protocolID)
	}
	if unitID := int(header[6]); unitID != s.config.UnitID {
		return nil, fmt.Errorf("unit ID mismatch: got %d, expected %d", unitID, s.config.UnitID)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("empty PDU")
	}

	responseFunction := body[0]
	if responseFunction == functionCode|0x80 {
		if len(body) < 2 {
			return nil, fmt.Errorf("Modbus exception response without exception code")
		}
		return nil, &ExceptionError{Code: body[1]}
	}
	if responseFunction != functionCode {
		return nil, fmt.Errorf("function code mismatch: got %d, expected %d", responseFunction, functionCode)
	}
	return body[1:], nil
}

func (s *Service) reconnect() error {
	s.mu.Lock()
	req := s.config
	s.mu.Unlock()
	conn, err := dial(req)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.conn = conn
	s.mu.Unlock()
	return nil
}

func (s *Service) closeConnection() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != nil {
		_ = s.conn.Close()
	}
	s.conn = nil
}

func (s *Service) setLastError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err == nil {
		s.lastError = ""
		return
	}
	s.lastError = err.Error()
}

func (s *Service) statusLocked() ConnectionStatus {
	return ConnectionStatus{
		Connected:      s.conn != nil,
		Host:           s.config.Host,
		Port:           s.config.Port,
		UnitID:         s.config.UnitID,
		TimeoutMs:      s.config.TimeoutMs,
		Retries:        s.config.Retries,
		LastError:      s.lastError,
		LastResponseMs: s.lastResponseMs,
	}
}

func (s *Service) appendFrameLogLocked(direction string, frame []byte, durationMs float64) {
	log := FrameLog{
		ID:         s.nextLogID,
		Timestamp:  timestamp(time.Now()),
		Direction:  direction,
		FullHex:    hexBytes(frame),
		DurationMs: durationMs,
	}
	s.nextLogID++

	if len(frame) >= 8 {
		log.TransactionID = binary.BigEndian.Uint16(frame[0:2])
		log.ProtocolID = binary.BigEndian.Uint16(frame[2:4])
		log.Length = binary.BigEndian.Uint16(frame[4:6])
		log.UnitID = frame[6]
		log.FunctionCode = frame[7]
		if len(frame) > 8 {
			log.PayloadHex = hexBytes(frame[8:])
		}
		if log.FunctionCode&0x80 != 0 && len(frame) > 8 {
			log.ExceptionCode = exceptionName(frame[8])
		}
	}

	s.logs = append(s.logs, log)
	if len(s.logs) > 500 {
		s.logs = s.logs[len(s.logs)-500:]
	}
}

func dial(req ConnectRequest) (net.Conn, error) {
	address := net.JoinHostPort(req.Host, strconv.Itoa(req.Port))
	return net.DialTimeout("tcp", address, time.Duration(req.TimeoutMs)*time.Millisecond)
}

func buildFrame(transactionID uint16, unitID byte, pdu []byte) []byte {
	frame := make([]byte, 7+len(pdu))
	binary.BigEndian.PutUint16(frame[0:2], transactionID)
	binary.BigEndian.PutUint16(frame[2:4], 0)
	binary.BigEndian.PutUint16(frame[4:6], uint16(len(pdu)+1))
	frame[6] = unitID
	copy(frame[7:], pdu)
	return frame
}

func writeSingleRegisterPDU(req WriteSingleRegisterRequest) []byte {
	pdu := make([]byte, 5)
	pdu[0] = FunctionWriteSingleRegister
	binary.BigEndian.PutUint16(pdu[1:3], req.Address)
	binary.BigEndian.PutUint16(pdu[3:5], req.Value)
	return pdu
}

func writeSingleCoilPDU(req WriteSingleCoilRequest) []byte {
	pdu := make([]byte, 5)
	pdu[0] = FunctionWriteSingleCoil
	binary.BigEndian.PutUint16(pdu[1:3], req.Address)
	binary.BigEndian.PutUint16(pdu[3:5], coilWireValue(req.Value))
	return pdu
}

func writeMultipleCoilsPDU(req WriteMultipleCoilsRequest) []byte {
	quantity := uint16(len(req.Values))
	packed := packBits(req.Values)
	pdu := make([]byte, 6+len(packed))
	pdu[0] = FunctionWriteMultipleCoils
	binary.BigEndian.PutUint16(pdu[1:3], req.Address)
	binary.BigEndian.PutUint16(pdu[3:5], quantity)
	pdu[5] = byte(len(packed))
	copy(pdu[6:], packed)
	return pdu
}

func writeMultipleRegistersPDU(req WriteMultipleRegistersRequest) []byte {
	quantity := uint16(len(req.Values))
	pdu := make([]byte, 6+len(req.Values)*2)
	pdu[0] = FunctionWriteMultipleRegisters
	binary.BigEndian.PutUint16(pdu[1:3], req.Address)
	binary.BigEndian.PutUint16(pdu[3:5], quantity)
	pdu[5] = byte(len(req.Values) * 2)
	for i, value := range req.Values {
		offset := 6 + i*2
		binary.BigEndian.PutUint16(pdu[offset:offset+2], value)
	}
	return pdu
}

func validateWriteQuantityEcho(functionCode byte, response []byte, address uint16, quantity uint16) error {
	if len(response) != 4 {
		return fmt.Errorf("unexpected FC%02d response length %d", functionCode, len(response))
	}
	if binary.BigEndian.Uint16(response[0:2]) != address || binary.BigEndian.Uint16(response[2:4]) != quantity {
		return fmt.Errorf("FC%02d response echo does not match request", functionCode)
	}
	return nil
}

func packBits(values []bool) []byte {
	packed := make([]byte, (len(values)+7)/8)
	for i, value := range values {
		if value {
			packed[i/8] |= 1 << (i % 8)
		}
	}
	return packed
}

func coilWireValue(value bool) uint16 {
	if value {
		return 0xFF00
	}
	return 0x0000
}

func coilPreviewValue(value bool) uint16 {
	if value {
		return 1
	}
	return 0
}

func coilLabel(value bool) string {
	if value {
		return "ON"
	}
	return "OFF"
}

func coilLabels(values []bool) []string {
	labels := make([]string, 0, len(values))
	for _, value := range values {
		labels = append(labels, coilLabel(value))
	}
	return labels
}

func registerValueLabels(values []uint16) []string {
	labels := make([]string, 0, len(values))
	for _, value := range values {
		labels = append(labels, fmt.Sprintf("%d", value))
	}
	return labels
}

func registerValuesHex(values []uint16) string {
	if len(values) == 0 {
		return ""
	}
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, fmt.Sprintf("0x%04X", value))
	}
	return strings.Join(parts, ", ")
}

func hexBytes(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	out := make([]byte, 0, len(data)*3-1)
	const digits = "0123456789ABCDEF"
	for i, b := range data {
		if i > 0 {
			out = append(out, ' ')
		}
		out = append(out, digits[b>>4], digits[b&0x0F])
	}
	return string(out)
}

func exceptionName(code byte) string {
	switch code {
	case 0x01:
		return "01 Illegal Function"
	case 0x02:
		return "02 Illegal Data Address"
	case 0x03:
		return "03 Illegal Data Value"
	case 0x04:
		return "04 Server Device Failure"
	case 0x05:
		return "05 Acknowledge"
	case 0x06:
		return "06 Server Device Busy"
	case 0x08:
		return "08 Memory Parity Error"
	case 0x0A:
		return "0A Gateway Path Unavailable"
	case 0x0B:
		return "0B Gateway Target Device Failed To Respond"
	default:
		return fmt.Sprintf("%02X Unknown Exception", code)
	}
}
