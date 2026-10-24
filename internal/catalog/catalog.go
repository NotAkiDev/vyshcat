// Package catalog — список того, что консоль умеет создавать
package catalog

import "vyshkat/internal/domain"

// VehicleKind - вид транспорта в меню приёма.
type VehicleKind struct {
	Title string
	// Electric: Проверка на наличие батареи и расхода
	Electric bool
	New      func(domain.VehicleParams, domain.ElectricParams) (domain.Vehicle, error)
}

// ItemKind - тип вещи в меню постановки на баланс
type ItemKind struct {
	Title string
	// Powered: Проверка на суточный расход
	Powered bool
	New     func(name string, dailyKWh float64) (domain.Inventoried, error)
}

// Catalog - Каталог меню
type Catalog struct {
	Vehicles []VehicleKind
	Items    []ItemKind
}

// Default - Каталог по умолчанию
func Default() Catalog {
	return Catalog{
		Vehicles: []VehicleKind{
			{
				Title: "Велосипед",
				New: func(p domain.VehicleParams, _ domain.ElectricParams) (domain.Vehicle, error) {
					return asVehicle(domain.NewBicycle(p))
				},
			},
			{
				Title:    "Электросамокат",
				Electric: true,
				New: func(p domain.VehicleParams, e domain.ElectricParams) (domain.Vehicle, error) {
					return asVehicle(domain.NewScooter(p, e))
				},
			},
			{
				Title:    "Электровелосипед",
				Electric: true,
				New: func(p domain.VehicleParams, e domain.ElectricParams) (domain.Vehicle, error) {
					return asVehicle(domain.NewEBike(p, e))
				},
			},
		},
		Items: []ItemKind{
			{
				Title: "Шлем",
				New: func(name string, _ float64) (domain.Inventoried, error) {
					return asItem(domain.NewHelmet(name))
				},
			},
			{
				Title:   "Док-станция",
				Powered: true,
				New: func(name string, kwh float64) (domain.Inventoried, error) {
					return asItem(domain.NewDockStation(name, kwh))
				},
			},
			{
				Title:   "Зарядный шкаф",
				Powered: true,
				New: func(name string, kwh float64) (domain.Inventoried, error) {
					return asItem(domain.NewChargingCabinet(name, kwh))
				},
			},
		},
	}
}

// asVehicle и asItem решают проблему с nil в интерфейсе при возврате
func asVehicle[T domain.Vehicle](v T, err error) (domain.Vehicle, error) {
	if err != nil {
		return nil, err
	}
	return v, nil
}

func asItem[T domain.Inventoried](it T, err error) (domain.Inventoried, error) {
	if err != nil {
		return nil, err
	}
	return it, nil
}
