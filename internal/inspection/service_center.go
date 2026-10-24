// Package inspection - техосмотр. Решает, пускать ли транспорт в парк
package inspection

import (
	"errors"
	"fmt"

	"vyshkat/internal/domain"
)

// ErrInvalidRules возвращается, если порог по батарее вне 0–100%
var ErrInvalidRules = errors.New("минимальная ёмкость батареи должна быть от 0 до 100%")

// Rules - настройки техосмотра
type Rules struct {
	// MinBatteryHealth —-ниже этого процента батарея считается изношенной
	MinBatteryHealth int
}

// DefaultRules - Примерно с 70% заряда самокат перестаёт дотягивать до конца смены
func DefaultRules() Rules {
	return Rules{MinBatteryHealth: 70}
}

// ServiceCenter - сервис техосмотра
type ServiceCenter struct {
	rules Rules
}

// NewServiceCenter не даст создать техосмотр с бессмысленным порогом
func NewServiceCenter(rules Rules) (*ServiceCenter, error) {
	if rules.MinBatteryHealth < 0 || rules.MinBatteryHealth > 100 {
		return nil, ErrInvalidRules
	}
	return &ServiceCenter{rules: rules}, nil
}

// Inspect смотрит тормоза и раму у всех
func (s *ServiceCenter) Inspect(v domain.Vehicle) domain.Verdict {
	var reasons []string

	c := v.Condition()
	if !c.BrakesOK {
		reasons = append(reasons, "неисправны тормоза")
	}
	if !c.FrameOK {
		reasons = append(reasons, "повреждена рама")
	}
	if b, ok := v.(domain.BatteryPowered); ok && b.BatteryHealth() < s.rules.MinBatteryHealth {
		reasons = append(reasons, fmt.Sprintf("износ батареи: %d%% при минимуме %d%%",
			b.BatteryHealth(), s.rules.MinBatteryHealth))
	}

	if len(reasons) > 0 {
		return domain.Reject(reasons...)
	}
	return domain.Approve()
}
