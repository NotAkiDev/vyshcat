package domain

// Реализует виды транспорта

// Bicycle - обычный велосипед, без мотора и без батареи
type Bicycle struct {
	vehicle
}

// NewBicycle Создаёт велосипед, если параметры в порядке
func NewBicycle(p VehicleParams) (*Bicycle, error) {
	p, err := p.validated()
	if err != nil {
		return nil, err
	}
	return &Bicycle{
		name:       p.Name,
		simplicity: p.Simplicity,
		condition:  p.Condition,
	}, nil
}

// Category Возвращает Велосипед
func (*Bicycle) Category() string { return "Велосипед" }

// Scooter - электросамокат
type Scooter struct {
	vehicle
	electric
}

// NewScooter создаёт электросамокат, а e описывает батарею и расход
func NewScooter(p VehicleParams, e ElectricParams) (*Scooter, error) {
	p, err := p.validated()
	if err != nil {
		return nil, err
	}
	if err := e.validate(); err != nil {
		return nil, err
	}
	return &Scooter{
		name:          p.Name,
		simplicity:    p.Simplicity,
		condition:     p.Condition,
		dailyKWh:      e.DailyKWh,
		batteryHealth: e.BatteryHealth,
	}, nil
}

// Category возвращает Электросамокат
func (*Scooter) Category() string { return "Электросамокат" }

// EBike - электровелосипед. Устроен как самокат, отличается только видом
type EBike struct {
	vehicle
	electric
}

// NewEBike создаёт электровелосипед, а e описывает батарею и расход
func NewEBike(p VehicleParams, e ElectricParams) (*EBike, error) {
	p, err := p.validated()
	if err != nil {
		return nil, err
	}
	if err := e.validate(); err != nil {
		return nil, err
	}
	return &EBike{
		name:          p.Name,
		simplicity:    p.Simplicity,
		condition:     p.Condition,
		dailyKWh:      e.DailyKWh,
		batteryHealth: e.BatteryHealth,
	}, nil
}

// Category возвращает объект Электровелосипед.
func (*EBike) Category() string { return "Электровелосипед" }

// Если вид потеряет метод, сборка упадёт прямо тут
var (
	_ Vehicle        = (*Bicycle)(nil)
	_ Vehicle        = (*Scooter)(nil)
	_ BatteryPowered = (*Scooter)(nil)
	_ Vehicle        = (*EBike)(nil)
	_ BatteryPowered = (*EBike)(nil)
)
