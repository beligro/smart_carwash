package modbus

import (
	"time"

	"github.com/google/uuid"
)

// ModbusServiceInterface интерфейс для работы с Modbus
type ModbusServiceInterface interface {
	WriteCoil(boxID uuid.UUID, register string, value bool) error
	WriteLightCoil(boxID uuid.UUID, value bool) error
	WriteChemistryCoil(boxID uuid.UUID, value bool) error
	HandleModbusError(boxID uuid.UUID, operation string, sessionID uuid.UUID, err error) error
	TestConnection(boxID uuid.UUID) error
	TestCoil(boxID uuid.UUID, register string, value bool) error
	ExtendSessionTime(sessionID uuid.UUID, duration time.Duration) error
}
