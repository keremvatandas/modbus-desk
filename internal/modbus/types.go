package modbus

import "time"

const (
	FunctionReadCoils              = 1
	FunctionReadDiscreteInputs     = 2
	FunctionReadHoldingRegisters   = 3
	FunctionReadInputRegisters     = 4
	FunctionWriteSingleCoil        = 5
	FunctionWriteSingleRegister    = 6
	FunctionWriteMultipleCoils     = 15
	FunctionWriteMultipleRegisters = 16
)

const (
	DataTypeUInt16  = "uint16"
	DataTypeInt16   = "int16"
	DataTypeFloat32 = "float32"
	DataTypeHex     = "hex"
	DataTypeBinary  = "binary"
)

const (
	ByteOrderABCD = "ABCD"
	ByteOrderBADC = "BADC"
	ByteOrderCDAB = "CDAB"
	ByteOrderDCBA = "DCBA"
)

type ConnectRequest struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	UnitID    int    `json:"unitId"`
	TimeoutMs int    `json:"timeoutMs"`
	Retries   int    `json:"retries"`
}

type ConnectionStatus struct {
	Connected      bool    `json:"connected"`
	Host           string  `json:"host"`
	Port           int     `json:"port"`
	UnitID         int     `json:"unitId"`
	TimeoutMs      int     `json:"timeoutMs"`
	Retries        int     `json:"retries"`
	LastError      string  `json:"lastError"`
	LastResponseMs float64 `json:"lastResponseMs"`
}

type ReadRegistersRequest struct {
	FunctionCode int    `json:"functionCode"`
	Address      uint16 `json:"address"`
	Quantity     uint16 `json:"quantity"`
	DataType     string `json:"dataType"`
	ByteOrder    string `json:"byteOrder"`
	AddressBase  int    `json:"addressBase"`
}

type ReadRegistersResponse struct {
	FunctionCode int           `json:"functionCode"`
	Address      uint16        `json:"address"`
	Quantity     uint16        `json:"quantity"`
	Rows         []RegisterRow `json:"rows"`
	DurationMs   float64       `json:"durationMs"`
}

type ReadBitsRequest struct {
	FunctionCode int    `json:"functionCode"`
	Address      uint16 `json:"address"`
	Quantity     uint16 `json:"quantity"`
	AddressBase  int    `json:"addressBase"`
}

type ReadBitsResponse struct {
	FunctionCode int      `json:"functionCode"`
	Address      uint16   `json:"address"`
	Quantity     uint16   `json:"quantity"`
	Rows         []BitRow `json:"rows"`
	DurationMs   float64  `json:"durationMs"`
}

type BitRow struct {
	Address        uint16 `json:"address"`
	DisplayAddress int    `json:"displayAddress"`
	Raw            string `json:"raw"`
	Value          bool   `json:"value"`
	Timestamp      string `json:"timestamp"`
	Quality        string `json:"quality"`
}

type RegisterRow struct {
	Address        uint16   `json:"address"`
	DisplayAddress int      `json:"displayAddress"`
	RegisterCount  int      `json:"registerCount"`
	RawHex         string   `json:"rawHex"`
	UInt16         uint16   `json:"uint16"`
	Int16          int16    `json:"int16"`
	Float32        *float64 `json:"float32,omitempty"`
	Binary         string   `json:"binary"`
	Value          any      `json:"value"`
	DataType       string   `json:"dataType"`
	Timestamp      string   `json:"timestamp"`
	Quality        string   `json:"quality"`
}

type WriteSingleRegisterRequest struct {
	Address   uint16 `json:"address"`
	Value     uint16 `json:"value"`
	DataType  string `json:"dataType"`
	ByteOrder string `json:"byteOrder"`
}

type WriteSingleCoilRequest struct {
	Address uint16 `json:"address"`
	Value   bool   `json:"value"`
}

type WriteMultipleCoilsRequest struct {
	Address uint16 `json:"address"`
	Values  []bool `json:"values"`
}

type WriteMultipleRegistersRequest struct {
	Address uint16   `json:"address"`
	Values  []uint16 `json:"values"`
}

type WritePreview struct {
	FunctionCode int      `json:"functionCode"`
	Address      uint16   `json:"address"`
	Quantity     uint16   `json:"quantity"`
	Value        uint16   `json:"value"`
	Values       []string `json:"values"`
	Host         string   `json:"host"`
	Port         int      `json:"port"`
	UnitID       int      `json:"unitId"`
	RegisterHex  string   `json:"registerHex"`
	PDUHex       string   `json:"pduHex"`
	FrameHex     string   `json:"frameHex"`
}

type FrameLog struct {
	ID            int     `json:"id"`
	Timestamp     string  `json:"timestamp"`
	Direction     string  `json:"direction"`
	TransactionID uint16  `json:"transactionId"`
	ProtocolID    uint16  `json:"protocolId"`
	Length        uint16  `json:"length"`
	UnitID        byte    `json:"unitId"`
	FunctionCode  byte    `json:"functionCode"`
	PayloadHex    string  `json:"payloadHex"`
	FullHex       string  `json:"fullHex"`
	DurationMs    float64 `json:"durationMs"`
	ExceptionCode string  `json:"exceptionCode,omitempty"`
}

func timestamp(t time.Time) string {
	return t.Format("15:04:05.000")
}
