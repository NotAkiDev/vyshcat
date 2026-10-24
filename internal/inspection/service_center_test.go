package inspection_test

import (
	"errors"
	"strings"
	"testing"

	"vyshkat/internal/domain"
	"vyshkat/internal/inspection"
)

func scooter(t *testing.T, c domain.Condition, battery int) domain.Vehicle {
	t.Helper()
	s, err := domain.NewScooter(domain.VehicleParams{Name: "S", Simplicity: 7, Condition: c},
		domain.ElectricParams{DailyKWh: 0.5, BatteryHealth: battery})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func bicycle(t *testing.T, c domain.Condition) domain.Vehicle {
	t.Helper()
	b, err := domain.NewBicycle(domain.VehicleParams{Name: "B", Simplicity: 7, Condition: c})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestNewServiceCenterValidatesRules(t *testing.T) {
	for _, threshold := range []int{-1, 101} {
		_, err := inspection.NewServiceCenter(inspection.Rules{MinBatteryHealth: threshold})
		if !errors.Is(err, inspection.ErrInvalidRules) {
			t.Errorf("MinBatteryHealth=%d: err = %v", threshold, err)
		}
	}
	if _, err := inspection.NewServiceCenter(inspection.DefaultRules()); err != nil {
		t.Errorf("правила по умолчанию: %v", err)
	}
}

func TestServiceCenterInspect(t *testing.T) {
	ok := domain.Condition{BrakesOK: true, FrameOK: true}
	sc, err := inspection.NewServiceCenter(inspection.Rules{MinBatteryHealth: 70})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		vehicle     domain.Vehicle
		approved    bool
		wantReasons []string
	}{
		{"исправный велосипед", bicycle(t, ok), true, nil},
		{"велосипед без тормозов", bicycle(t, domain.Condition{FrameOK: true}), false, []string{"тормоза"}},
		{"самокат с битой рамой", scooter(t, domain.Condition{BrakesOK: true}, 90), false, []string{"рама"}},
		{"батарея ровно на пороге", scooter(t, ok, 70), true, nil},
		{"батарея ниже порога", scooter(t, ok, 69), false, []string{"износ батареи: 69%"}},
		{"всё плохо — все причины", scooter(t, domain.Condition{}, 10), false,
			[]string{"тормоза", "рама", "батареи"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := sc.Inspect(tt.vehicle)
			if v.Approved != tt.approved {
				t.Fatalf("Approved = %v, want %v (%v)", v.Approved, tt.approved, v.Reasons)
			}
			if len(v.Reasons) != len(tt.wantReasons) {
				t.Fatalf("причины %v, ожидалось %d шт.", v.Reasons, len(tt.wantReasons))
			}
			for i, want := range tt.wantReasons {
				if !strings.Contains(v.Reasons[i], want) {
					t.Errorf("причина %d = %q, want содержит %q", i, v.Reasons[i], want)
				}
			}
		})
	}
}
