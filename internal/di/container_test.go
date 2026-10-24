package di_test

import (
	"bytes"
	"strings"
	"testing"

	"go.uber.org/dig"

	"vyshkat/internal/console"
	"vyshkat/internal/di"
	"vyshkat/internal/domain"
	"vyshkat/internal/fleet"
	"vyshkat/internal/inspection"
)

func newContainer(t *testing.T) *dig.Container {
	t.Helper()
	c, err := di.NewContainer(di.DefaultConfig(strings.NewReader(""), &bytes.Buffer{}))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// fakeInspector Заменяет контейнер настоящего техосмотра, считает сколько раз к нему обратились
type fakeInspector struct {
	verdict domain.Verdict
	calls   int
}

func (f *fakeInspector) Inspect(domain.Vehicle) domain.Verdict {
	f.calls++
	return f.verdict
}

// substitute подменяет fleet.Inspector через Decorate
func substitute(t *testing.T, c *dig.Container, fake fleet.Inspector) {
	t.Helper()
	if err := c.Decorate(func(fleet.Inspector) fleet.Inspector { return fake }); err != nil {
		t.Fatal(err)
	}
}

func resolveFleet(t *testing.T, c *dig.Container) *fleet.Fleet {
	t.Helper()
	var f *fleet.Fleet
	if err := c.Invoke(func(resolved *fleet.Fleet) { f = resolved }); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestContainerBuildsApp(t *testing.T) {
	c := newContainer(t)
	if err := c.Invoke(func(*console.App) {}); err != nil {
		t.Fatalf("приложение не собирается: %v", err)
	}
}

func TestContainerRegistersServiceCenterAsInspector(t *testing.T) {
	c := newContainer(t)
	err := c.Invoke(func(i fleet.Inspector) {
		if _, ok := i.(*inspection.ServiceCenter); !ok {
			t.Errorf("fleet.Inspector = %T, want *inspection.ServiceCenter", i)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestContainerProvidesSingleFleet(t *testing.T) {
	c := newContainer(t)
	if resolveFleet(t, c) != resolveFleet(t, c) {
		t.Error("контейнер создал два разных парка")
	}
}

func TestContainerRejectsInvalidConfig(t *testing.T) {
	cfg := di.DefaultConfig(strings.NewReader(""), &bytes.Buffer{})
	cfg.Inspection.MinBatteryHealth = 150
	c, err := di.NewContainer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Invoke(func(*fleet.Fleet) {}); err == nil {
		t.Error("ожидалась ошибка сборки при некорректных порогах техосмотра")
	}
}

// Отказывает в принятии идеально исправному велосипеду
func TestSubstituteInspector_Reject(t *testing.T) {
	c := newContainer(t)
	fake := &fakeInspector{verdict: domain.Reject("техосмотр закрыт на санитарный день")}
	substitute(t, c, fake)
	f := resolveFleet(t, c)

	b, err := domain.NewBicycle(domain.VehicleParams{Name: "Исправный", Simplicity: 9,
		Condition: domain.Condition{BrakesOK: true, FrameOK: true}})
	if err != nil {
		t.Fatal(err)
	}
	verdict, err := f.AcceptVehicle(b)
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Approved || fake.calls != 1 {
		t.Errorf("verdict = %+v, вызовов fake = %d", verdict, fake.calls)
	}
	if f.Summary().Total() != 0 {
		t.Error("отклонённый транспорт попал в парк")
	}
}

// Пропускает самокат с батареей на 10%
func TestSubstituteInspector_Approve(t *testing.T) {
	c := newContainer(t)
	fake := &fakeInspector{verdict: domain.Approve()}
	substitute(t, c, fake)
	f := resolveFleet(t, c)

	worn, err := domain.NewScooter(domain.VehicleParams{Name: "Изношенный", Simplicity: 7,
		Condition: domain.Condition{BrakesOK: true, FrameOK: true}},
		domain.ElectricParams{DailyKWh: 0.5, BatteryHealth: 10})
	if err != nil {
		t.Fatal(err)
	}
	verdict, err := f.AcceptVehicle(worn)
	if err != nil {
		t.Fatal(err)
	}
	if !verdict.Approved || worn.InventoryNumber() != 1 {
		t.Errorf("verdict = %+v, номер = %s", verdict, worn.InventoryNumber())
	}
}
