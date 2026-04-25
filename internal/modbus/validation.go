package modbus

import (
	"errors"
	"fmt"
	"strings"
)

func NormalizeConnectRequest(req ConnectRequest) ConnectRequest {
	req.Host = strings.TrimSpace(req.Host)
	if req.Port == 0 {
		req.Port = 502
	}
	if req.TimeoutMs == 0 {
		req.TimeoutMs = 3000
	}
	if req.Retries < 0 {
		req.Retries = 0
	}
	return req
}

func ValidateConnectRequest(req ConnectRequest) error {
	req = NormalizeConnectRequest(req)
	if req.Host == "" {
		return errors.New("host/IP is required")
	}
	if req.Port < 1 || req.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if req.UnitID < 0 || req.UnitID > 247 {
		return fmt.Errorf("unit ID must be between 0 and 247")
	}
	if req.TimeoutMs < 100 || req.TimeoutMs > 120000 {
		return fmt.Errorf("timeout must be between 100ms and 120000ms")
	}
	if req.Retries < 0 || req.Retries > 10 {
		return fmt.Errorf("retry count must be between 0 and 10")
	}
	return nil
}

func ValidateReadRegistersRequest(req ReadRegistersRequest) error {
	if req.FunctionCode != FunctionReadHoldingRegisters && req.FunctionCode != FunctionReadInputRegisters {
		return fmt.Errorf("function code %d is not supported for register reads", req.FunctionCode)
	}
	if req.Quantity == 0 || req.Quantity > 125 {
		return fmt.Errorf("quantity must be between 1 and 125")
	}
	if err := validateAddressRange(req.Address, req.Quantity); err != nil {
		return err
	}
	if err := ValidateDataType(req.DataType); err != nil {
		return err
	}
	if err := ValidateByteOrder(req.ByteOrder); err != nil {
		return err
	}
	if req.AddressBase != 0 && req.AddressBase != 1 {
		return fmt.Errorf("address base must be 0 or 1")
	}
	return nil
}

func ValidateReadBitsRequest(req ReadBitsRequest) error {
	if req.FunctionCode != FunctionReadCoils && req.FunctionCode != FunctionReadDiscreteInputs {
		return fmt.Errorf("function code %d is not supported for bit reads", req.FunctionCode)
	}
	if req.Quantity == 0 || req.Quantity > 2000 {
		return fmt.Errorf("quantity must be between 1 and 2000")
	}
	if err := validateAddressRange(req.Address, req.Quantity); err != nil {
		return err
	}
	if req.AddressBase != 0 && req.AddressBase != 1 {
		return fmt.Errorf("address base must be 0 or 1")
	}
	return nil
}

func ValidateWriteSingleRegisterRequest(req WriteSingleRegisterRequest) error {
	if err := validateAddressRange(req.Address, 1); err != nil {
		return err
	}
	if req.DataType != "" && req.DataType != DataTypeUInt16 && req.DataType != DataTypeHex {
		return fmt.Errorf("FC06 MVP writes support uint16 values only")
	}
	if req.ByteOrder != "" {
		return ValidateByteOrder(req.ByteOrder)
	}
	return nil
}

func ValidateWriteSingleCoilRequest(req WriteSingleCoilRequest) error {
	return validateAddressRange(req.Address, 1)
}

func ValidateWriteMultipleCoilsRequest(req WriteMultipleCoilsRequest) error {
	quantity := len(req.Values)
	if quantity < 1 || quantity > 1968 {
		return fmt.Errorf("coil quantity must be between 1 and 1968")
	}
	return validateAddressRange(req.Address, uint16(quantity))
}

func ValidateWriteMultipleRegistersRequest(req WriteMultipleRegistersRequest) error {
	quantity := len(req.Values)
	if quantity < 1 || quantity > 123 {
		return fmt.Errorf("register quantity must be between 1 and 123")
	}
	return validateAddressRange(req.Address, uint16(quantity))
}

func ValidateDataType(dataType string) error {
	switch normalizeDataType(dataType) {
	case DataTypeUInt16, DataTypeInt16, DataTypeFloat32, DataTypeHex, DataTypeBinary:
		return nil
	default:
		return fmt.Errorf("unsupported data type %q", dataType)
	}
}

func ValidateByteOrder(order string) error {
	switch normalizeByteOrder(order) {
	case ByteOrderABCD, ByteOrderBADC, ByteOrderCDAB, ByteOrderDCBA:
		return nil
	default:
		return fmt.Errorf("unsupported byte order %q", order)
	}
}

func normalizeDataType(dataType string) string {
	switch strings.ToLower(strings.TrimSpace(dataType)) {
	case "", "uint16", "u16":
		return DataTypeUInt16
	case "int16", "i16":
		return DataTypeInt16
	case "float32", "float", "f32":
		return DataTypeFloat32
	case "hex":
		return DataTypeHex
	case "binary", "bin":
		return DataTypeBinary
	default:
		return strings.ToLower(strings.TrimSpace(dataType))
	}
}

func normalizeByteOrder(order string) string {
	switch strings.ToUpper(strings.TrimSpace(order)) {
	case "":
		return ByteOrderABCD
	default:
		return strings.ToUpper(strings.TrimSpace(order))
	}
}

func validateAddressRange(address uint16, quantity uint16) error {
	if int(address)+int(quantity) > 65536 {
		return fmt.Errorf("address plus quantity exceeds Modbus address range")
	}
	return nil
}
