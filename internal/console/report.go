package console

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"text/tabwriter"

	"vyshkat/internal/domain"
)

// tabwriter выравнивает таблицы

func (a *App) printSummary() {
	s := a.fleet.Summary()
	fmt.Fprintln(a.out, "Отчёт по парку")
	w := tabwriter.NewWriter(a.out, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  Транспорт\t%d\n", s.Vehicles)
	fmt.Fprintf(w, "  Вещи\t%d\n", s.Items)
	fmt.Fprintf(w, "  Всего единиц\t%d\n", s.Total())
	if len(s.ByCategory) > 0 {
		fmt.Fprintln(w, "  По видам:\t")
		for _, c := range slices.Sorted(maps.Keys(s.ByCategory)) {
			fmt.Fprintf(w, "    %s\t%d\n", c, s.ByCategory[c])
		}
	}
	fmt.Fprintln(w, "  Суточное энергопотребление:\t")
	fmt.Fprintf(w, "    транспорт\t%s\n", formatKWh(s.VehicleEnergyKWh))
	fmt.Fprintf(w, "    вещи\t%s\n", formatKWh(s.ItemEnergyKWh))
	fmt.Fprintf(w, "    всего\t%s\n", formatKWh(s.DailyEnergyKWh()))
	w.Flush()
}

func (a *App) printBeginners() {
	rule := a.fleet.BeginnerRule()
	fmt.Fprintf(a.out, "Транспорт для новичков (простота ≥ %d)\n", rule.MinSimplicity)
	list := a.fleet.BeginnerFriendly()
	if len(list) == 0 {
		fmt.Fprintln(a.out, "  Подходящего транспорта нет.")
		return
	}
	w := tabwriter.NewWriter(a.out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "  Инв. №\tВид\tНаименование\tПростота")
	for _, v := range list {
		fmt.Fprintf(w, "  %s\t%s\t%s\t%d\n", v.InventoryNumber(), v.Category(), v.Name(), v.Simplicity())
	}
	w.Flush()
}

func (a *App) printInventory() {
	fmt.Fprintln(a.out, "Инвентарная ведомость")
	list := a.fleet.Inventory()
	if len(list) == 0 {
		fmt.Fprintln(a.out, "  На балансе ничего нет.")
		return
	}
	w := tabwriter.NewWriter(a.out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "  Инв. №\tКатегория\tНаименование\tРасход в сутки")
	for _, it := range list {
		// Прочерк, а не ноль: шлем не «тратит 0», он вообще ничего не тратит.
		energy := "—"
		if c, ok := it.(domain.EnergyConsumer); ok {
			energy = formatKWh(c.DailyEnergyKWh())
		}
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n", it.InventoryNumber(), it.Category(), it.Name(), energy)
	}
	w.Flush()
}

// formatKWh Форматирует вывод киловатт
func formatKWh(x float64) string {
	return strings.Replace(fmt.Sprintf("%.2f кВт·ч", x), ".", ",", 1)
}
