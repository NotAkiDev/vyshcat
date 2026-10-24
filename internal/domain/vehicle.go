package domain

import "errors"

// Шкала простоты для новичка
const (
	MinSimplicity = 1
	MaxSimplicity = 10
)

// Ошибки проверки параметров транспорта
var (
	ErrSimplicityRange = errors.New("простота для новичка должна быть от 1 до 10")
	ErrBatteryRange    = errors.New("остаточная ёмкость батареи должна быть от 0 до 100%")
)

// Vehicle - то, что выдаётся в прокат
type Vehicle interface {
	Inventoried
	// Simplicity - насколько легко на этом ездить
	Simplicity() int
	// Condition - то, на что смотрит техосмотр
	Condition() Condition
}

// BatteryPowered - транспорт с батареей
type BatteryPowered interface {
	EnergyConsumer
	// BatteryHealth - сколько процентов ёмкости осталось от заводской
	BatteryHealth() int
}

// Condition - состояние узлов у любого транспорта
type Condition struct {
	BrakesOK bool
	FrameOK  bool
}

// VehicleParams собирает в одну структуру то, что общее у всех видов
type VehicleParams struct {
	Name       string
	Simplicity int
	Condition  Condition
}

// ElectricParams - то, что есть только у электротранспорта
type ElectricParams struct {
	DailyKWh      float64
	BatteryHealth int
}

// vehicle встраивается во все виды транспорта
type vehicle struct {
	asset
	simplicity int
	condition  Condition
}

// validated проверяет параметры и чистит имя от пробелов по краям
func (p VehicleParams) validated() (VehicleParams, error) {
	name, err := validName(p.Name)
	if err != nil {
		return VehicleParams{}, err
	}
	if p.Simplicity < MinSimplicity || p.Simplicity > MaxSimplicity {
		return VehicleParams{}, ErrSimplicityRange
	}
	p.Name = name
	return p, nil
}

func (v *vehicle) Simplicity() int { return v.simplicity }

func (v *vehicle) Condition() Condition { return v.condition }

// electric встраивается в электросамокат и электровелосипед
type electric struct {
	powered
	batteryHealth int
}

func (e ElectricParams) validate() error {
	if err := validateEnergy(e.DailyKWh); err != nil {
		return err
	}
	if e.BatteryHealth < 0 || e.BatteryHealth > 100 {
		return ErrBatteryRange
	}
	return nil
}

func (e *electric) BatteryHealth() int { return e.batteryHealth }
