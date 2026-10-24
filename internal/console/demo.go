package console

import (
	"fmt"

	"vyshkat/internal/domain"
)

// RunDemo показывает все сценарии подряд
func (a *App) RunDemo() {
	ok := domain.Condition{BrakesOK: true, FrameOK: true}
	noBrakes := domain.Condition{BrakesOK: false, FrameOK: true}
	brokenFrame := domain.Condition{BrakesOK: true, FrameOK: false}
	params := func(name string, simplicity int, c domain.Condition) domain.VehicleParams {
		return domain.VehicleParams{Name: name, Simplicity: simplicity, Condition: c}
	}
	battery := func(kwh float64, health int) domain.ElectricParams {
		return domain.ElectricParams{DailyKWh: kwh, BatteryHealth: health}
	}

	a.step("0. Состояние парка до демо")
	a.printReports()

	a.step("1. Приём транспорта через техосмотр")
	a.acceptOrReport(domain.NewBicycle(params("Stels Navigator 300", 9, ok)))
	ninebot, err := domain.NewScooter(params("Ninebot Max G30", 7, ok), battery(0.55, 92))
	a.acceptOrReport(ninebot, err)
	a.acceptOrReport(domain.NewEBike(params("Eltreco XT 800", 4, ok), battery(0.9, 85)))
	a.acceptOrReport(domain.NewScooter(params("Kugoo S3", 6, ok), battery(0.35, 78)))

	a.step("2. Отказы техосмотра")
	a.acceptOrReport(domain.NewScooter(params("Xiaomi Pro 2", 8, noBrakes), battery(0.45, 90)))
	a.acceptOrReport(domain.NewEBike(params("Wenbox U2", 6, brokenFrame), battery(0.8, 40)))

	a.step("3. Повторный приём уже принятого транспорта")
	if ninebot != nil {
		a.accept(ninebot)
	}

	a.step("4. Некорректные данные отклоняются при создании")
	a.acceptOrReport(domain.NewBicycle(params("Forward Racing", 11, ok)))
	a.registerOrReport(domain.NewDockStation("Док-станция с отрицательным расходом", -1))

	a.step("5. Вещи на балансе")
	a.registerOrReport(domain.NewHelmet("Шлем детский S"))
	a.registerOrReport(domain.NewHelmet("Шлем взрослый L"))
	a.registerOrReport(domain.NewDockStation("Док-станция у корпуса на Покровке", 1.2))
	a.registerOrReport(domain.NewChargingCabinet("Зарядный шкаф на 12 ячеек", 3.6))

	a.step("6. Транспорт в обход техосмотра не ставится на баланс")
	if bypass, err := domain.NewBicycle(params("Велосипед с заднего двора", 7, ok)); err == nil {
		a.register(bypass)
	}

	a.step("7. Итоговые отчёты")
	a.printReports()
}

func (a *App) step(title string) {
	fmt.Fprintf(a.out, "\n=== %s ===\n", title)
}

func (a *App) printReports() {
	a.printSummary()
	fmt.Fprintln(a.out)
	a.printBeginners()
	fmt.Fprintln(a.out)
	a.printInventory()
}

// acceptOrReport и registerOrReport принимают сразу пару объект, ошибка
func (a *App) acceptOrReport(v domain.Vehicle, err error) {
	if err != nil {
		fmt.Fprintf(a.out, "  [ошибка]   %v\n", err)
		return
	}
	a.accept(v)
}

func (a *App) registerOrReport(item domain.Inventoried, err error) {
	if err != nil {
		fmt.Fprintf(a.out, "  [ошибка]   %v\n", err)
		return
	}
	a.register(item)
}
