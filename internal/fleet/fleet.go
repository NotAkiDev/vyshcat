// Package fleet ведёт учёт парка
package fleet

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"vyshkat/internal/domain"
)

// Ошибки учёта
var (
	ErrNilInspector           = errors.New("не задан сервис техосмотра")
	ErrInvalidBeginnerRule    = errors.New("порог простоты для новичков должен быть от 1 до 10")
	ErrNilAsset               = errors.New("объект не задан")
	ErrAlreadyRegistered      = errors.New("объект уже на балансе")
	ErrVehicleNeedsInspection = errors.New("транспорт ставится на баланс только через техосмотр")
)

type Inspector interface {
	Inspect(v domain.Vehicle) domain.Verdict
}

// BeginnerRule решает, можно ли дать транспорт новичку
type BeginnerRule struct {
	// MinSimplicity — минимальная простота
	MinSimplicity int
}

// DefaultBeginnerRule По умолчанию, простота от 6 и выше
func DefaultBeginnerRule() BeginnerRule {
	return BeginnerRule{MinSimplicity: 6}
}

// Allows отвечает, подходит ли транспорт новичку
func (r BeginnerRule) Allows(v domain.Vehicle) bool {
	return v.Simplicity() >= r.MinSimplicity
}

// Fleet - сам парк
type Fleet struct {
	inspector  Inspector
	beginners  BeginnerRule
	vehicles   []domain.Vehicle
	items      []domain.Inventoried
	lastNumber domain.InventoryNumber
}

// New создаёт пустой парк
func New(inspector Inspector, beginners BeginnerRule) (*Fleet, error) {
	if inspector == nil {
		return nil, ErrNilInspector
	}
	if beginners.MinSimplicity < domain.MinSimplicity || beginners.MinSimplicity > domain.MaxSimplicity {
		return nil, ErrInvalidBeginnerRule
	}
	return &Fleet{inspector: inspector, beginners: beginners}, nil
}

// AcceptVehicle отправляет транспорт на техосмотр и, если его одобрили, выдаёт номер и ставит на баланс
func (f *Fleet) AcceptVehicle(v domain.Vehicle) (domain.Verdict, error) {
	if v == nil {
		return domain.Verdict{}, ErrNilAsset
	}
	if n := v.InventoryNumber(); n != 0 {
		return domain.Verdict{}, fmt.Errorf("%w: %s", ErrAlreadyRegistered, n)
	}

	verdict := f.inspector.Inspect(v)
	if !verdict.Approved {
		return verdict, nil
	}
	if err := f.assignNumber(v); err != nil {
		return domain.Verdict{}, err
	}
	f.vehicles = append(f.vehicles, v)
	return verdict, nil
}

// RegisterItem ставит на баланс вещь
func (f *Fleet) RegisterItem(item domain.Inventoried) error {
	if item == nil {
		return ErrNilAsset
	}
	if _, isVehicle := item.(domain.Vehicle); isVehicle {
		return ErrVehicleNeedsInspection
	}
	if err := f.assignNumber(item); err != nil {
		return err
	}
	f.items = append(f.items, item)
	return nil
}

// assignNumber выдаёт следующий инвентарный номер
func (f *Fleet) assignNumber(a domain.Inventoried) error {
	if n := a.InventoryNumber(); n != 0 {
		return fmt.Errorf("%w: %s", ErrAlreadyRegistered, n)
	}
	next := f.lastNumber + 1
	if err := a.AssignInventoryNumber(next); err != nil {
		return err
	}
	f.lastNumber = next
	return nil
}

// BeginnerRule нужен консоли, чтобы показать порог в заголовке отчёта
func (f *Fleet) BeginnerRule() BeginnerRule { return f.beginners }

// Inventory - общая ведомость
func (f *Fleet) Inventory() []domain.Inventoried {
	all := make([]domain.Inventoried, 0, len(f.vehicles)+len(f.items))
	for _, v := range f.vehicles {
		all = append(all, v)
	}
	all = append(all, f.items...)
	slices.SortFunc(all, func(a, b domain.Inventoried) int {
		return cmp.Compare(a.InventoryNumber(), b.InventoryNumber())
	})
	return all
}

// TotalDailyEnergyKWh - Потребление за сутки
func (f *Fleet) TotalDailyEnergyKWh() float64 {
	return dailyEnergyKWh(f.vehicles) + dailyEnergyKWh(f.items)
}

// dailyEnergyKWh складывает расход тех, кто реализует EnergyConsumer
func dailyEnergyKWh[T domain.Inventoried](list []T) float64 {
	var total float64
	for _, a := range list {
		if c, ok := any(a).(domain.EnergyConsumer); ok {
			total += c.DailyEnergyKWh()
		}
	}
	return total
}

// BeginnerFriendly - транспорт, который можно выдавать новичкам
func (f *Fleet) BeginnerFriendly() []domain.Vehicle {
	var res []domain.Vehicle
	for _, v := range f.vehicles {
		if f.beginners.Allows(v) {
			res = append(res, v)
		}
	}
	return res
}

// Summary - сводка по парку для отчёта в консоли
type Summary struct {
	Vehicles         int
	Items            int
	ByCategory       map[string]int
	VehicleEnergyKWh float64
	ItemEnergyKWh    float64
}

// Total - сколько всего единиц на балансе
func (s Summary) Total() int { return s.Vehicles + s.Items }

// DailyEnergyKWh - общий расход за сутки
func (s Summary) DailyEnergyKWh() float64 { return s.VehicleEnergyKWh + s.ItemEnergyKWh }

// Summary считает, сколько чего на балансе и сколько это потребляет
func (f *Fleet) Summary() Summary {
	byCategory := make(map[string]int)
	for _, a := range f.Inventory() {
		byCategory[a.Category()]++
	}
	return Summary{
		Vehicles:         len(f.vehicles),
		Items:            len(f.items),
		ByCategory:       byCategory,
		VehicleEnergyKWh: dailyEnergyKWh(f.vehicles),
		ItemEnergyKWh:    dailyEnergyKWh(f.items),
	}
}
