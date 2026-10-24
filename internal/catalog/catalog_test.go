package catalog_test

import (
	"testing"

	"vyshkat/internal/catalog"
	"vyshkat/internal/domain"
)

// Электросамокат создаёт именно электросамокат
// Флаг Electric стоит ровно у тех, у кого есть батарея
func TestVehicleKinds(t *testing.T) {
	p := domain.VehicleParams{Name: "Модель", Simplicity: 6, Condition: domain.Condition{BrakesOK: true, FrameOK: true}}
	e := domain.ElectricParams{DailyKWh: 0.5, BatteryHealth: 90}
	for _, k := range catalog.Default().Vehicles {
		t.Run(k.Title, func(t *testing.T) {
			v, err := k.New(p, e)
			if err != nil {
				t.Fatal(err)
			}
			if v.Category() != k.Title {
				t.Errorf("Category() = %q, want %q", v.Category(), k.Title)
			}
			if _, isBattery := v.(domain.BatteryPowered); isBattery != k.Electric {
				t.Errorf("BatteryPowered = %v, Electric = %v", isBattery, k.Electric)
			}
		})
	}
}

// При ошибке должен вернуться nil.
func TestVehicleKinds_ErrorReturnsNil(t *testing.T) {
	for _, k := range catalog.Default().Vehicles {
		v, err := k.New(domain.VehicleParams{Name: "", Simplicity: 5}, domain.ElectricParams{})
		if err == nil || v != nil {
			t.Errorf("%s: v = %v, err = %v", k.Title, v, err)
		}
	}
}

func TestItemKinds(t *testing.T) {
	for _, k := range catalog.Default().Items {
		t.Run(k.Title, func(t *testing.T) {
			it, err := k.New("Вещь", 1.5)
			if err != nil {
				t.Fatal(err)
			}
			if it.Category() != k.Title {
				t.Errorf("Category() = %q, want %q", it.Category(), k.Title)
			}
			if _, isEnergy := it.(domain.EnergyConsumer); isEnergy != k.Powered {
				t.Errorf("EnergyConsumer = %v, Powered = %v", isEnergy, k.Powered)
			}
			if bad, err := k.New("", 1); err == nil || bad != nil {
				t.Errorf("пустое имя: v = %v, err = %v", bad, err)
			}
		})
	}
}
