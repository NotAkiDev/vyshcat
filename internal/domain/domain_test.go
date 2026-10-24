package domain_test

import (
	"errors"
	"math"
	"testing"

	"vyshkat/internal/domain"
)

var okCondition = domain.Condition{BrakesOK: true, FrameOK: true}

func TestInventoryNumberString(t *testing.T) {
	if got := domain.InventoryNumber(42).String(); got != "ВК-000042" {
		t.Errorf("String() = %q, want ВК-000042", got)
	}
	if got := domain.InventoryNumber(0).String(); got != "без номера" {
		t.Errorf("String() = %q, want «без номера»", got)
	}
}

func TestAssignInventoryNumber(t *testing.T) {
	h, err := domain.NewHelmet("Шлем M")
	if err != nil {
		t.Fatal(err)
	}
	if h.InventoryNumber() != 0 {
		t.Fatalf("новый объект уже с номером %s", h.InventoryNumber())
	}
	if err := h.AssignInventoryNumber(0); !errors.Is(err, domain.ErrInvalidNumber) {
		t.Errorf("номер 0: err = %v, want ErrInvalidNumber", err)
	}
	if err := h.AssignInventoryNumber(7); err != nil {
		t.Fatalf("первое присвоение: %v", err)
	}
	if err := h.AssignInventoryNumber(8); !errors.Is(err, domain.ErrAlreadyNumbered) {
		t.Errorf("повторное присвоение: err = %v, want ErrAlreadyNumbered", err)
	}
	if h.InventoryNumber() != 7 {
		t.Errorf("номер изменился на %s", h.InventoryNumber())
	}
}

func TestVehicleValidation(t *testing.T) {
	vp := func(name string, simplicity int) domain.VehicleParams {
		return domain.VehicleParams{Name: name, Simplicity: simplicity}
	}
	ep := func(kwh float64, battery int) domain.ElectricParams {
		return domain.ElectricParams{DailyKWh: kwh, BatteryHealth: battery}
	}
	tests := []struct {
		name    string
		params  domain.VehicleParams
		e       domain.ElectricParams
		wantErr error
	}{
		{"пустое имя", vp("  ", 5), ep(0.5, 90), domain.ErrEmptyName},
		{"простота 0", vp("X", 0), ep(0.5, 90), domain.ErrSimplicityRange},
		{"простота 11", vp("X", 11), ep(0.5, 90), domain.ErrSimplicityRange},
		{"отрицательный расход", vp("X", 5), ep(-0.1, 90), domain.ErrInvalidEnergy},
		{"расход NaN", vp("X", 5), ep(math.NaN(), 90), domain.ErrInvalidEnergy},
		{"расход больше максимума", vp("X", 5), ep(domain.MaxDailyKWh+1, 90), domain.ErrInvalidEnergy},
		{"батарея 101%", vp("X", 5), ep(0.5, 101), domain.ErrBatteryRange},
		{"батарея -1%", vp("X", 5), ep(0.5, -1), domain.ErrBatteryRange},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := domain.NewScooter(tt.params, tt.e); !errors.Is(err, tt.wantErr) {
				t.Errorf("NewScooter: err = %v, want %v", err, tt.wantErr)
			}
			if _, err := domain.NewEBike(tt.params, tt.e); !errors.Is(err, tt.wantErr) {
				t.Errorf("NewEBike: err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestBicycleValidationAndFields(t *testing.T) {
	_, err := domain.NewBicycle(domain.VehicleParams{Name: "", Simplicity: 5})
	if !errors.Is(err, domain.ErrEmptyName) {
		t.Errorf("пустое имя: err = %v", err)
	}
	_, err = domain.NewBicycle(domain.VehicleParams{Name: "B", Simplicity: 12})
	if !errors.Is(err, domain.ErrSimplicityRange) {
		t.Errorf("простота 12: err = %v", err)
	}
	b, err := domain.NewBicycle(domain.VehicleParams{Name: "  Stels  ", Simplicity: 10, Condition: okCondition})
	if err != nil {
		t.Fatal(err)
	}
	if b.Name() != "Stels" || b.Simplicity() != 10 || b.Condition() != okCondition || b.Category() != "Велосипед" {
		t.Errorf("поля велосипеда: %q %d %+v %q", b.Name(), b.Simplicity(), b.Condition(), b.Category())
	}
}

func TestItemValidation(t *testing.T) {
	if _, err := domain.NewHelmet(""); !errors.Is(err, domain.ErrEmptyName) {
		t.Errorf("шлем без имени: err = %v", err)
	}
	if _, err := domain.NewDockStation("", 1); !errors.Is(err, domain.ErrEmptyName) {
		t.Errorf("док-станция без имени: err = %v", err)
	}
	if _, err := domain.NewDockStation("D", math.Inf(1)); !errors.Is(err, domain.ErrInvalidEnergy) {
		t.Errorf("док-станция с бесконечным расходом: err = %v", err)
	}
	if _, err := domain.NewChargingCabinet("C", -3); !errors.Is(err, domain.ErrInvalidEnergy) {
		t.Errorf("шкаф с отрицательным расходом: err = %v", err)
	}
	c, err := domain.NewChargingCabinet("C", 3.6)
	if err != nil {
		t.Fatal(err)
	}
	if c.DailyEnergyKWh() != 3.6 || c.Category() != "Зарядный шкаф" {
		t.Errorf("поля шкафа: %v %q", c.DailyEnergyKWh(), c.Category())
	}
}

// У каждого типа ровно те возможности, которые нужны
func TestCapabilities(t *testing.T) {
	vp := domain.VehicleParams{Name: "V", Simplicity: 5}
	ep := domain.ElectricParams{DailyKWh: 0.5, BatteryHealth: 80}
	helmet, errH := domain.NewHelmet("H")
	bike, errB := domain.NewBicycle(vp)
	scooter, errS := domain.NewScooter(vp, ep)
	ebike, errE := domain.NewEBike(vp, ep)
	dock, errD := domain.NewDockStation("D", 1)
	cabinet, errC := domain.NewChargingCabinet("C", 2)
	if err := errors.Join(errH, errB, errS, errE, errD, errC); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		obj                      domain.Inventoried
		category                 string
		vehicle, energy, battery bool
	}{
		{helmet, "Шлем", false, false, false},
		{bike, "Велосипед", true, false, false},
		{scooter, "Электросамокат", true, true, true},
		{ebike, "Электровелосипед", true, true, true},
		{dock, "Док-станция", false, true, false},
		{cabinet, "Зарядный шкаф", false, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			if got := tt.obj.Category(); got != tt.category {
				t.Errorf("Category() = %q", got)
			}
			_, isVehicle := tt.obj.(domain.Vehicle)
			_, isEnergy := tt.obj.(domain.EnergyConsumer)
			_, isBattery := tt.obj.(domain.BatteryPowered)
			if isVehicle != tt.vehicle || isEnergy != tt.energy || isBattery != tt.battery {
				t.Errorf("vehicle=%v energy=%v battery=%v, want %v %v %v",
					isVehicle, isEnergy, isBattery, tt.vehicle, tt.energy, tt.battery)
			}
		})
	}
	if scooter.BatteryHealth() != 80 || scooter.DailyEnergyKWh() != 0.5 {
		t.Errorf("поля самоката: %d %v", scooter.BatteryHealth(), scooter.DailyEnergyKWh())
	}
}

func TestVerdict(t *testing.T) {
	if v := domain.Approve(); !v.Approved || len(v.Reasons) != 0 {
		t.Errorf("Approve() = %+v", v)
	}
	if v := domain.Reject("a", "b"); v.Approved || len(v.Reasons) != 2 {
		t.Errorf("Reject() = %+v", v)
	}
}
