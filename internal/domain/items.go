package domain

// Вещи числятся на балансе, но в прокат не выдаются и техосмотр не проходят.

// Helmet - шлем
type Helmet struct {
	asset
}

// NewHelmet создаёт шлем
func NewHelmet(name string) (*Helmet, error) {
	name, err := validName(name)
	if err != nil {
		return nil, err
	}
	return &Helmet{name: name}, nil
}

// Category возвращает шлем.
func (*Helmet) Category() string { return "Шлем" }

// DockStation - док-станция, где самокаты паркуются и заряжаются
type DockStation struct {
	asset
	powered
}

// NewDockStation создаёт док-станцию с расходом dailyKWh за сутки
func NewDockStation(name string, dailyKWh float64) (*DockStation, error) {
	name, err := validPoweredItem(name, dailyKWh)
	if err != nil {
		return nil, err
	}
	return &DockStation{name: name, dailyKWh: dailyKWh}, nil
}

// Category возвращает объект Док-станция;
func (*DockStation) Category() string { return "Док-станция" }

// ChargingCabinet - шкаф, в котором заряжаются съёмные батареи
type ChargingCabinet struct {
	asset
	powered
}

// NewChargingCabinet создаёт зарядный шкаф с расходом dailyKWh за сутки
func NewChargingCabinet(name string, dailyKWh float64) (*ChargingCabinet, error) {
	name, err := validPoweredItem(name, dailyKWh)
	if err != nil {
		return nil, err
	}
	return &ChargingCabinet{name: name, dailyKWh: dailyKWh}, nil
}

// Category возвращает объект Зарядный шкаф.
func (*ChargingCabinet) Category() string { return "Зарядный шкаф" }

// validPoweredItem общий для док-станции и шкафа
func validPoweredItem(name string, dailyKWh float64) (string, error) {
	name, err := validName(name)
	if err != nil {
		return "", err
	}
	if err := validateEnergy(dailyKWh); err != nil {
		return "", err
	}
	return name, nil
}

// Если какой-то тип перестанет реализовывать нужный интерфейс, сборка упадёт здесь
var (
	_ Inventoried    = (*Helmet)(nil)
	_ Inventoried    = (*DockStation)(nil)
	_ EnergyConsumer = (*DockStation)(nil)
	_ Inventoried    = (*ChargingCabinet)(nil)
	_ EnergyConsumer = (*ChargingCabinet)(nil)
)
