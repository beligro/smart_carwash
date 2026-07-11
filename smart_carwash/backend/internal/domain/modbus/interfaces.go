package modbus

import (
	"context"

	"github.com/google/uuid"
)

// ModbusServiceInterface интерфейс для работы с Modbus
type ModbusServiceInterface interface {
	WriteCoil(ctx context.Context, boxID uuid.UUID, register string, value bool) error
	WriteLightCoil(ctx context.Context, boxID uuid.UUID, register string, value bool) error
	WriteChemistryCoil(ctx context.Context, boxID uuid.UUID, register string, value bool) error
	HandleModbusError(boxID uuid.UUID, operation string, sessionID uuid.UUID, err error) error
	TestCoil(ctx context.Context, boxID uuid.UUID, register string, value bool) error
	// GetCoilStatus возвращает последний известный статус коилов (света/химии) бокса.
	// Если статус ещё не сохранён — возвращает (nil, nil, nil).
	GetCoilStatus(ctx context.Context, boxID uuid.UUID) (light *bool, chem *bool, err error)
}
