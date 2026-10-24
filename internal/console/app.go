// Package console Меню, через которое пользователь работает с парком
package console

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"vyshkat/internal/catalog"
	"vyshkat/internal/domain"
	"vyshkat/internal/fleet"
)

// IO - откуда читаем и куда пишем
type IO struct {
	In  io.Reader
	Out io.Writer
}

// App - консольное приложение поверх парка и каталога
type App struct {
	fleet   *fleet.Fleet
	catalog catalog.Catalog
	ask     prompter
	out     io.Writer
}

// New собирает приложение, выдывает DI-контейнер
func New(f *fleet.Fleet, c catalog.Catalog, stdio IO) *App {
	return &App{
		fleet:   f,
		catalog: c,
		ask:     prompter{in: bufio.NewReader(stdio.In), out: stdio.Out},
		out:     stdio.Out,
	}
}

type menuItem struct {
	title  string
	action func() error
}

func (a *App) menu() []menuItem {
	return []menuItem{
		{"Принять транспорт (с техосмотром)", a.acceptVehicle},
		{"Поставить вещь на баланс", a.registerItem},
		{"Отчёт: число единиц и энергопотребление", a.noInput(a.printSummary)},
		{"Транспорт для новичков", a.noInput(a.printBeginners)},
		{"Инвентарная ведомость", a.noInput(a.printInventory)},
		{"Демо-сценарий", a.noInput(a.RunDemo)},
	}
}

// noInput подгоняет отчёты под общий вид пункта меню
func (a *App) noInput(f func()) func() error {
	return func() error { f(); return nil }
}

// Run Отвечает за вызов меню, пока пользователь не выберет выход.
func (a *App) Run() error {
	fmt.Fprintln(a.out, "ВышКат — учёт парка. «Доедем до дедлайна»")
	items := a.menu()
	for {
		fmt.Fprintln(a.out)
		for i, it := range items {
			fmt.Fprintf(a.out, "%d. %s\n", i+1, it.title)
		}
		fmt.Fprintln(a.out, "0. Выход")

		choice, err := a.ask.intInRange("Пункт меню", 0, len(items))
		if err == nil && choice == 0 {
			fmt.Fprintln(a.out, "До встречи!")
			return nil
		}
		if err == nil {
			fmt.Fprintln(a.out)
			err = items[choice-1].action()
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func (a *App) acceptVehicle() error {
	kinds := a.catalog.Vehicles
	fmt.Fprintln(a.out, "Вид транспорта:")
	for i, k := range kinds {
		fmt.Fprintf(a.out, "  %d. %s\n", i+1, k.Title)
	}
	idx, err := a.ask.intInRange("Номер вида", 1, len(kinds))
	if err != nil {
		return err
	}
	kind := kinds[idx-1]

	var p domain.VehicleParams
	if p.Name, err = a.ask.text("Наименование (модель)"); err != nil {
		return err
	}
	if p.Simplicity, err = a.ask.intInRange("Простота для новичка (1–10)",
		domain.MinSimplicity, domain.MaxSimplicity); err != nil {
		return err
	}
	if p.Condition.BrakesOK, err = a.ask.yesNo("Тормоза исправны?"); err != nil {
		return err
	}
	if p.Condition.FrameOK, err = a.ask.yesNo("Рама без повреждений?"); err != nil {
		return err
	}

	var e domain.ElectricParams
	if kind.Electric {
		if e.DailyKWh, err = a.ask.floatInRange("Суточный расход энергии, кВт·ч", 0, domain.MaxDailyKWh); err != nil {
			return err
		}
		if e.BatteryHealth, err = a.ask.intInRange("Остаточная ёмкость батареи, %", 0, 100); err != nil {
			return err
		}
	}

	v, err := kind.New(p, e)
	if err != nil {
		fmt.Fprintf(a.out, "  Не удалось оформить транспорт: %v\n", err)
		return nil
	}
	a.accept(v)
	return nil
}

func (a *App) registerItem() error {
	kinds := a.catalog.Items
	fmt.Fprintln(a.out, "Вид вещи:")
	for i, k := range kinds {
		fmt.Fprintf(a.out, "  %d. %s\n", i+1, k.Title)
	}
	idx, err := a.ask.intInRange("Номер вида", 1, len(kinds))
	if err != nil {
		return err
	}
	kind := kinds[idx-1]

	name, err := a.ask.text("Наименование")
	if err != nil {
		return err
	}
	var kwh float64
	if kind.Powered {
		if kwh, err = a.ask.floatInRange("Суточный расход энергии, кВт·ч", 0, domain.MaxDailyKWh); err != nil {
			return err
		}
	}

	item, err := kind.New(name, kwh)
	if err != nil {
		fmt.Fprintf(a.out, "  Не удалось оформить вещь: %v\n", err)
		return nil
	}
	a.register(item)
	return nil
}

// accept и register печатают, чем закончилась операция
func (a *App) accept(v domain.Vehicle) {
	verdict, err := a.fleet.AcceptVehicle(v)
	switch {
	case err != nil:
		fmt.Fprintf(a.out, "  [ошибка]   %s «%s»: %v\n", v.Category(), v.Name(), err)
	case verdict.Approved:
		fmt.Fprintf(a.out, "  [принят]   %s «%s», инв. № %s\n", v.Category(), v.Name(), v.InventoryNumber())
	default:
		fmt.Fprintf(a.out, "  [отклонён] %s «%s»: %s\n", v.Category(), v.Name(), strings.Join(verdict.Reasons, "; "))
	}
}

func (a *App) register(item domain.Inventoried) {
	if err := a.fleet.RegisterItem(item); err != nil {
		fmt.Fprintf(a.out, "  [ошибка]   %s «%s»: %v\n", item.Category(), item.Name(), err)
		return
	}
	fmt.Fprintf(a.out, "  [на балансе] %s «%s», инв. № %s\n", item.Category(), item.Name(), item.InventoryNumber())
}
