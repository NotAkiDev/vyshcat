package fleet_test

import (
	"errors"
	"math"
	"testing"

	"vyshkat/internal/domain"
	"vyshkat/internal/fleet"
)

// stubInspector всегда отвечает одно и то же и записывает, что ему приносили
type stubInspector struct {
	verdict   domain.Verdict
	inspected []domain.Vehicle
}

func (s *stubInspector) Inspect(v domain.Vehicle) domain.Verdict {
	s.inspected = append(s.inspected, v)
	return s.verdict
}

func approveAll() *stubInspector { return &stubInspector{verdict: domain.Approve()} }

// inspectorFunc нужен, когда решение зависит от транспорта
type inspectorFunc func(domain.Vehicle) domain.Verdict

func (f inspectorFunc) Inspect(v domain.Vehicle) domain.Verdict { return f(v) }

func newFleet(t *testing.T, insp fleet.Inspector) *fleet.Fleet {
	t.Helper()
	f, err := fleet.New(insp, fleet.DefaultBeginnerRule())
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func bike(t *testing.T, name string, simplicity int) *domain.Bicycle {
	t.Helper()
	b, err := domain.NewBicycle(domain.VehicleParams{Name: name, Simplicity: simplicity,
		Condition: domain.Condition{BrakesOK: true, FrameOK: true}})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func scooter(t *testing.T, name string, simplicity int, kwh float64) *domain.Scooter {
	t.Helper()
	s, err := domain.NewScooter(domain.VehicleParams{Name: name, Simplicity: simplicity,
		Condition: domain.Condition{BrakesOK: true, FrameOK: true}},
		domain.ElectricParams{DailyKWh: kwh, BatteryHealth: 90})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestNewValidatesDependencies(t *testing.T) {
	if _, err := fleet.New(nil, fleet.DefaultBeginnerRule()); !errors.Is(err, fleet.ErrNilInspector) {
		t.Errorf("без техосмотра: err = %v", err)
	}
	for _, threshold := range []int{0, 11} {
		_, err := fleet.New(approveAll(), fleet.BeginnerRule{MinSimplicity: threshold})
		if !errors.Is(err, fleet.ErrInvalidBeginnerRule) {
			t.Errorf("MinSimplicity=%d: err = %v", threshold, err)
		}
	}
}

// Положительный ответ проведения техосмотра
func TestAcceptVehicle_Approved(t *testing.T) {
	insp := approveAll()
	f := newFleet(t, insp)
	b := bike(t, "Stels", 8)

	verdict, err := f.AcceptVehicle(b)
	if err != nil {
		t.Fatal(err)
	}
	if !verdict.Approved {
		t.Fatalf("verdict = %+v", verdict)
	}
	if len(insp.inspected) != 1 || insp.inspected[0] != b {
		t.Errorf("техосмотр вызван %d раз", len(insp.inspected))
	}
	if b.InventoryNumber() != 1 {
		t.Errorf("номер = %s, want ВК-000001", b.InventoryNumber())
	}
	if inv := f.Inventory(); len(inv) != 1 || inv[0] != b {
		t.Errorf("Inventory() = %v", inv)
	}
}

// Отказ послпе техосмотра
func TestAcceptVehicle_Rejected(t *testing.T) {
	insp := &stubInspector{verdict: domain.Reject("неисправны тормоза")}
	f := newFleet(t, insp)
	b := bike(t, "Stels", 8)

	verdict, err := f.AcceptVehicle(b)
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Approved || len(verdict.Reasons) != 1 || verdict.Reasons[0] != "неисправны тормоза" {
		t.Errorf("verdict = %+v", verdict)
	}
	if b.InventoryNumber() != 0 {
		t.Errorf("отклонённому транспорту присвоен номер %s", b.InventoryNumber())
	}
	if s := f.Summary(); s.Total() != 0 {
		t.Errorf("в парке %d единиц, want 0", s.Total())
	}
	if len(f.BeginnerFriendly()) != 0 {
		t.Error("отклонённый транспорт попал в список для новичков")
	}
}

func TestAcceptVehicle_Twice(t *testing.T) {
	insp := approveAll()
	f := newFleet(t, insp)
	b := bike(t, "Stels", 8)
	if _, err := f.AcceptVehicle(b); err != nil {
		t.Fatal(err)
	}
	if _, err := f.AcceptVehicle(b); !errors.Is(err, fleet.ErrAlreadyRegistered) {
		t.Errorf("повторный приём: err = %v", err)
	}
	if len(insp.inspected) != 1 {
		t.Errorf("повторный приём снова отправил на техосмотр (%d вызовов)", len(insp.inspected))
	}
}

func TestAcceptVehicle_Nil(t *testing.T) {
	f := newFleet(t, approveAll())
	if _, err := f.AcceptVehicle(nil); !errors.Is(err, fleet.ErrNilAsset) {
		t.Errorf("err = %v", err)
	}
}

// Один транспорт был отбракован, один принят
func TestAcceptVehicle_DecisionPerVehicle(t *testing.T) {
	f := newFleet(t, inspectorFunc(func(v domain.Vehicle) domain.Verdict {
		if v.Name() == "Брак" {
			return domain.Reject("брак")
		}
		return domain.Approve()
	}))
	good, bad := bike(t, "Годный", 7), bike(t, "Брак", 7)
	mustAccept(t, f, good)
	mustAccept(t, f, bad)
	if inv := f.Inventory(); len(inv) != 1 || inv[0] != good {
		t.Errorf("Inventory() = %v", inv)
	}
}

func TestRegisterItem(t *testing.T) {
	f := newFleet(t, approveAll())
	b := bike(t, "Stels", 8)
	mustAccept(t, f, b)

	h := helmet(t, "Шлем L")
	if err := f.RegisterItem(h); err != nil {
		t.Fatal(err)
	}
	if h.InventoryNumber() != 2 {
		t.Errorf("сквозная нумерация: номер шлема %s, want ВК-000002", h.InventoryNumber())
	}
	if err := f.RegisterItem(h); !errors.Is(err, fleet.ErrAlreadyRegistered) {
		t.Errorf("повторная постановка: err = %v", err)
	}
	if err := f.RegisterItem(nil); !errors.Is(err, fleet.ErrNilAsset) {
		t.Errorf("nil: err = %v", err)
	}
}

// Велосипед пытаются провести как вещь
func TestRegisterItem_RejectsVehicle(t *testing.T) {
	insp := approveAll()
	f := newFleet(t, insp)
	b := bike(t, "Stels", 8)
	if err := f.RegisterItem(b); !errors.Is(err, fleet.ErrVehicleNeedsInspection) {
		t.Errorf("err = %v, want ErrVehicleNeedsInspection", err)
	}
	if b.InventoryNumber() != 0 || f.Summary().Total() != 0 {
		t.Error("транспорт попал на баланс в обход техосмотра")
	}
}

// Энергию считаем по всему балансу
func TestTotalDailyEnergy(t *testing.T) {
	f := newFleet(t, approveAll())
	mustAccept(t, f, bike(t, "Велосипед", 8))
	mustAccept(t, f, scooter(t, "Самокат", 7, 0.55))
	d, errD := domain.NewDockStation("Док", 1.2)
	c, errC := domain.NewChargingCabinet("Шкаф", 3.6)
	if err := errors.Join(errD, errC); err != nil {
		t.Fatal(err)
	}
	for _, it := range []domain.Inventoried{helmet(t, "Шлем"), d, c} {
		mustRegister(t, f, it)
	}

	if got, want := f.TotalDailyEnergyKWh(), 0.55+1.2+3.6; !almostEqual(got, want) {
		t.Errorf("TotalDailyEnergyKWh() = %v, want %v", got, want)
	}
	s := f.Summary()
	if !almostEqual(s.VehicleEnergyKWh, 0.55) || !almostEqual(s.ItemEnergyKWh, 4.8) {
		t.Errorf("разбивка энергии: транспорт %v, вещи %v", s.VehicleEnergyKWh, s.ItemEnergyKWh)
	}
	if !almostEqual(s.DailyEnergyKWh(), f.TotalDailyEnergyKWh()) {
		t.Errorf("Summary.DailyEnergyKWh() = %v", s.DailyEnergyKWh())
	}
	if got := newFleet(t, approveAll()).TotalDailyEnergyKWh(); got != 0 {
		t.Errorf("пустой парк: %v", got)
	}
}

// Простота 6 проходит, 5 нет
func TestBeginnerFriendly(t *testing.T) {
	f := newFleet(t, approveAll())
	s5 := bike(t, "Простота 5", 5)
	s6 := scooter(t, "Простота 6", 6, 0.4)
	s10 := bike(t, "Простота 10", 10)
	for _, v := range []domain.Vehicle{s5, s6, s10} {
		mustAccept(t, f, v)
	}

	got := f.BeginnerFriendly()
	if len(got) != 2 || got[0] != s6 || got[1] != s10 {
		t.Errorf("BeginnerFriendly() = %v, want [Простота 6, Простота 10]", names(got))
	}
}

func TestBeginnerFriendly_CustomRule(t *testing.T) {
	f, err := fleet.New(approveAll(), fleet.BeginnerRule{MinSimplicity: 8})
	if err != nil {
		t.Fatal(err)
	}
	s7, s9 := bike(t, "7", 7), bike(t, "9", 9)
	mustAccept(t, f, s7)
	mustAccept(t, f, s9)
	if got := f.BeginnerFriendly(); len(got) != 1 || got[0] != s9 {
		t.Errorf("BeginnerFriendly() = %v", names(got))
	}
	if f.BeginnerRule().MinSimplicity != 8 {
		t.Errorf("BeginnerRule() = %+v", f.BeginnerRule())
	}
}

func TestSummaryAndInventoryOrder(t *testing.T) {
	f := newFleet(t, approveAll())
	mustRegister(t, f, helmet(t, "Шлем 1"))
	mustAccept(t, f, scooter(t, "Самокат 1", 7, 0.5))
	mustRegister(t, f, helmet(t, "Шлем 2"))
	mustAccept(t, f, scooter(t, "Самокат 2", 7, 0.5))
	mustAccept(t, f, bike(t, "Велосипед", 7))

	s := f.Summary()
	if s.Vehicles != 3 || s.Items != 2 || s.Total() != 5 {
		t.Errorf("Summary = %+v", s)
	}
	want := map[string]int{"Электросамокат": 2, "Велосипед": 1, "Шлем": 2}
	for k, v := range want {
		if s.ByCategory[k] != v {
			t.Errorf("ByCategory[%s] = %d, want %d", k, s.ByCategory[k], v)
		}
	}
	if !almostEqual(s.VehicleEnergyKWh, 1.0) || s.ItemEnergyKWh != 0 {
		t.Errorf("энергия: транспорт %v, вещи %v", s.VehicleEnergyKWh, s.ItemEnergyKWh)
	}

	inv := f.Inventory()
	for i, it := range inv {
		if it.InventoryNumber() != domain.InventoryNumber(i+1) {
			t.Errorf("позиция %d: номер %s", i, it.InventoryNumber())
		}
	}
}

func helmet(t *testing.T, name string) *domain.Helmet {
	t.Helper()
	h, err := domain.NewHelmet(name)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func almostEqual(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func mustAccept(t *testing.T, f *fleet.Fleet, v domain.Vehicle) {
	t.Helper()
	if _, err := f.AcceptVehicle(v); err != nil {
		t.Fatal(err)
	}
}

func mustRegister(t *testing.T, f *fleet.Fleet, it domain.Inventoried) {
	t.Helper()
	if err := f.RegisterItem(it); err != nil {
		t.Fatal(err)
	}
}

func names(vs []domain.Vehicle) []string {
	res := make([]string, len(vs))
	for i, v := range vs {
		res[i] = v.Name()
	}
	return res
}
