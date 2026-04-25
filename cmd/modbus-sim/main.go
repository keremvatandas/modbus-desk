package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
)

const (
	fcReadCoils              = 1
	fcReadDiscreteInputs     = 2
	fcReadHoldingRegisters   = 3
	fcReadInputRegisters     = 4
	fcWriteSingleCoil        = 5
	fcWriteSingleRegister    = 6
	fcWriteMultipleCoils     = 15
	fcWriteMultipleRegisters = 16
)

type registerBank struct {
	mu      sync.Mutex
	coils   []bool
	inputs  []bool
	holding []uint16
	inputsR []uint16
}

func main() {
	host := flag.String("host", "127.0.0.1", "listen host")
	port := flag.Int("port", 1502, "listen port")
	flag.Parse()

	bank := registerBank{
		coils:   make([]bool, 256),
		inputs:  make([]bool, 256),
		holding: make([]uint16, 256),
		inputsR: make([]uint16, 256),
	}
	seedRegisters(&bank)

	listener, err := net.Listen("tcp", net.JoinHostPort(*host, strconv.Itoa(*port)))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Printf("ModbusDesk simulator listening on %s", listener.Addr())
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go handleConn(conn, &bank)
	}
}

func seedRegisters(bank *registerBank) {
	bank.coils[0] = true
	bank.coils[2] = true
	bank.coils[3] = true
	bank.coils[7] = true
	bank.coils[8] = true

	bank.inputs[1] = true
	bank.inputs[4] = true
	bank.inputs[5] = true
	bank.inputs[6] = true

	bank.holding[0] = 0x42F6
	bank.holding[1] = 0x0000
	bank.holding[2] = 0x4120
	bank.holding[3] = 0x0000
	bank.holding[10] = 230

	bank.inputsR[0] = 0x42C8
	bank.inputsR[1] = 0x0000
	bank.inputsR[2] = 0x4248
	bank.inputsR[3] = 0x0000
}

func handleConn(conn net.Conn, bank *registerBank) {
	defer conn.Close()
	log.Printf("client connected: %s", conn.RemoteAddr())

	for {
		header := make([]byte, 7)
		if _, err := io.ReadFull(conn, header); err != nil {
			log.Printf("client disconnected: %s", conn.RemoteAddr())
			return
		}
		length := int(binary.BigEndian.Uint16(header[4:6]))
		if length < 2 {
			return
		}
		body := make([]byte, length-1)
		if _, err := io.ReadFull(conn, body); err != nil {
			log.Printf("read body: %v", err)
			return
		}

		txID := binary.BigEndian.Uint16(header[0:2])
		unitID := header[6]
		pdu := handlePDU(body, bank)
		if _, err := conn.Write(buildFrame(txID, unitID, pdu)); err != nil {
			log.Printf("write response: %v", err)
			return
		}
	}
}

func handlePDU(body []byte, bank *registerBank) []byte {
	if len(body) == 0 {
		return []byte{0x80, 0x03}
	}

	functionCode := body[0]
	switch functionCode {
	case fcReadCoils, fcReadDiscreteInputs:
		if len(body) < 5 {
			return exception(functionCode, 0x03)
		}
		address := int(binary.BigEndian.Uint16(body[1:3]))
		quantity := int(binary.BigEndian.Uint16(body[3:5]))
		return readBits(functionCode, address, quantity, bank)
	case fcReadHoldingRegisters, fcReadInputRegisters:
		if len(body) < 5 {
			return exception(functionCode, 0x03)
		}
		address := int(binary.BigEndian.Uint16(body[1:3]))
		quantity := int(binary.BigEndian.Uint16(body[3:5]))
		return readRegisters(functionCode, address, quantity, bank)
	case fcWriteSingleCoil, fcWriteSingleRegister:
		if len(body) < 5 {
			return exception(functionCode, 0x03)
		}
		address := int(binary.BigEndian.Uint16(body[1:3]))
		value := binary.BigEndian.Uint16(body[3:5])
		if functionCode == fcWriteSingleCoil {
			return writeSingleCoil(functionCode, address, value, bank)
		}
		return writeSingleRegister(functionCode, address, value, bank)
	case fcWriteMultipleCoils:
		if len(body) < 6 {
			return exception(functionCode, 0x03)
		}
		address := int(binary.BigEndian.Uint16(body[1:3]))
		quantity := int(binary.BigEndian.Uint16(body[3:5]))
		byteCount := int(body[5])
		if len(body) < 6+byteCount {
			return exception(functionCode, 0x03)
		}
		return writeMultipleCoils(functionCode, address, quantity, body[6:6+byteCount], bank)
	case fcWriteMultipleRegisters:
		if len(body) < 6 {
			return exception(functionCode, 0x03)
		}
		address := int(binary.BigEndian.Uint16(body[1:3]))
		quantity := int(binary.BigEndian.Uint16(body[3:5]))
		byteCount := int(body[5])
		if len(body) < 6+byteCount {
			return exception(functionCode, 0x03)
		}
		return writeMultipleRegisters(functionCode, address, quantity, body[6:6+byteCount], bank)
	default:
		return exception(functionCode, 0x01)
	}
}

func readRegisters(functionCode byte, address int, quantity int, bank *registerBank) []byte {
	if quantity < 1 || quantity > 125 {
		return exception(functionCode, 0x03)
	}

	bank.mu.Lock()
	defer bank.mu.Unlock()

	registers := bank.holding
	if functionCode == fcReadInputRegisters {
		registers = bank.inputsR
	}
	if address < 0 || address+quantity > len(registers) {
		return exception(functionCode, 0x02)
	}

	pdu := make([]byte, 2+quantity*2)
	pdu[0] = functionCode
	pdu[1] = byte(quantity * 2)
	for i := 0; i < quantity; i++ {
		binary.BigEndian.PutUint16(pdu[2+i*2:4+i*2], registers[address+i])
	}
	return pdu
}

func readBits(functionCode byte, address int, quantity int, bank *registerBank) []byte {
	if quantity < 1 || quantity > 2000 {
		return exception(functionCode, 0x03)
	}

	bank.mu.Lock()
	defer bank.mu.Unlock()

	bits := bank.coils
	if functionCode == fcReadDiscreteInputs {
		bits = bank.inputs
	}
	if address < 0 || address+quantity > len(bits) {
		return exception(functionCode, 0x02)
	}

	byteCount := (quantity + 7) / 8
	pdu := make([]byte, 2+byteCount)
	pdu[0] = functionCode
	pdu[1] = byte(byteCount)
	for i := 0; i < quantity; i++ {
		if bits[address+i] {
			pdu[2+i/8] |= 1 << uint(i%8)
		}
	}
	return pdu
}

func writeSingleRegister(functionCode byte, address int, value uint16, bank *registerBank) []byte {
	bank.mu.Lock()
	defer bank.mu.Unlock()

	if address < 0 || address >= len(bank.holding) {
		return exception(functionCode, 0x02)
	}
	bank.holding[address] = value
	log.Printf("holding[%d] = %d (0x%04X)", address, value, value)

	pdu := make([]byte, 5)
	pdu[0] = functionCode
	binary.BigEndian.PutUint16(pdu[1:3], uint16(address))
	binary.BigEndian.PutUint16(pdu[3:5], value)
	return pdu
}

func writeSingleCoil(functionCode byte, address int, value uint16, bank *registerBank) []byte {
	if value != 0xFF00 && value != 0x0000 {
		return exception(functionCode, 0x03)
	}

	bank.mu.Lock()
	defer bank.mu.Unlock()

	if address < 0 || address >= len(bank.coils) {
		return exception(functionCode, 0x02)
	}
	bank.coils[address] = value == 0xFF00
	log.Printf("coil[%d] = %t", address, bank.coils[address])

	pdu := make([]byte, 5)
	pdu[0] = functionCode
	binary.BigEndian.PutUint16(pdu[1:3], uint16(address))
	binary.BigEndian.PutUint16(pdu[3:5], value)
	return pdu
}

func writeMultipleCoils(functionCode byte, address int, quantity int, data []byte, bank *registerBank) []byte {
	if quantity < 1 || quantity > 1968 || len(data) != (quantity+7)/8 {
		return exception(functionCode, 0x03)
	}

	bank.mu.Lock()
	defer bank.mu.Unlock()

	if address < 0 || address+quantity > len(bank.coils) {
		return exception(functionCode, 0x02)
	}
	for i := 0; i < quantity; i++ {
		bank.coils[address+i] = data[i/8]&(1<<uint(i%8)) != 0
	}
	log.Printf("coils[%d:%d] updated", address, address+quantity)

	return writeQuantityEcho(functionCode, address, quantity)
}

func writeMultipleRegisters(functionCode byte, address int, quantity int, data []byte, bank *registerBank) []byte {
	if quantity < 1 || quantity > 123 || len(data) != quantity*2 {
		return exception(functionCode, 0x03)
	}

	bank.mu.Lock()
	defer bank.mu.Unlock()

	if address < 0 || address+quantity > len(bank.holding) {
		return exception(functionCode, 0x02)
	}
	for i := 0; i < quantity; i++ {
		offset := i * 2
		bank.holding[address+i] = binary.BigEndian.Uint16(data[offset : offset+2])
	}
	log.Printf("holding[%d:%d] updated", address, address+quantity)

	return writeQuantityEcho(functionCode, address, quantity)
}

func writeQuantityEcho(functionCode byte, address int, quantity int) []byte {
	pdu := make([]byte, 5)
	pdu[0] = functionCode
	binary.BigEndian.PutUint16(pdu[1:3], uint16(address))
	binary.BigEndian.PutUint16(pdu[3:5], uint16(quantity))
	return pdu
}

func exception(functionCode byte, code byte) []byte {
	return []byte{functionCode | 0x80, code}
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

func init() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: modbus-sim [-host 127.0.0.1] [-port 1502]\n")
		flag.PrintDefaults()
	}
}
