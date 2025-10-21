package modbus

import (
	"github.com/google/uuid"
)

// ModbusServiceInterface интерфейс для работы с Modbus
type ModbusServiceInterface interface {
	WriteCoil(boxID uuid.UUID, register string, value bool) error
	WriteLightCoil(boxID uuid.UUID, value bool) error
	WriteChemistryCoil(boxID uuid.UUID, value bool) error
	HandleModbusError(boxID uuid.UUID, operation string, sessionID uuid.UUID, err error) error
	TestCoil(boxID uuid.UUID, register string, value bool) error
}
