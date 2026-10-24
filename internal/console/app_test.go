package console_test

import (
	"bytes"
	"strings"
	"testing"

	"vyshkat/internal/catalog"
	"vyshkat/internal/console"
	"vyshkat/internal/fleet"
	"vyshkat/internal/inspection"
)

// Проверка стандартного сценария ввода
func run(t *testing.T, input string) string {
	t.Helper()
	sc, err := inspection.NewServiceCenter(inspection.DefaultRules())
	if err != nil {
		t.Fatal(err)
	}
	f, err := fleet.New(sc, fleet.DefaultBeginnerRule())
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	app := console.New(f, catalog.Default(), console.IO{In: strings.NewReader(input), Out: &out})
	if err := app.Run(); err != nil {
		t.Fatalf("Run() = %v\n%s", err, out.String())
	}
	return out.String()
}

func assertContains(t *testing.T, out string, parts ...string) {
	t.Helper()
	for _, p := range parts {
		if !strings.Contains(out, p) {
			t.Errorf("в выводе нет %q\n--- вывод ---\n%s", p, out)
		}
	}
}

func lines(ls ...string) string { return strings.Join(ls, "\n") + "\n" }

func TestAcceptScooterAndReports(t *testing.T) {
	out := run(t, lines(
		"1",               // принять транспорт
		"2",               // электросамокат
		"Ninebot Max G30", // наименование
		"7",               // простота
		"д", "да",         // тормоза, рама
		"0,55",        // расход
		"92",          // батарея
		"3", "4", "5", // отчёт, новички, ведомость
		"0", // выход
	))
	assertContains(t, out,
		"[принят]   Электросамокат «Ninebot Max G30», инв. № ВК-000001",
		"Всего единиц",
		"0,55 кВт·ч",
		"Транспорт для новичков (простота ≥ 6)",
		"ВК-000001  Электросамокат  Ninebot Max G30",
		"До встречи!",
	)
}

func TestRejectedByInspection(t *testing.T) {
	out := run(t, lines("1", "1", "Stels", "9", "н", "д", "5", "0"))
	assertContains(t, out,
		"[отклонён] Велосипед «Stels»: неисправны тормоза",
		"На балансе ничего нет.",
	)
}

func TestInvalidInputIsReAsked(t *testing.T) {
	out := run(t, lines(
		"abc", "9", // некорректные пункты меню
		"1", "1",
		"",        // пустое имя
		"Forward", // имя
		"15", "8", // простота вне диапазона
		"может", "y", "n", // невалидный ответ
		"0",
	))
	assertContains(t, out,
		"Введите целое число от 0 до 6.",
		"Значение не может быть пустым.",
		"Введите целое число от 1 до 10.",
		"Ответьте «д» или «н».",
		"[отклонён] Велосипед «Forward»: повреждена рама",
	)
}

func TestRegisterItems(t *testing.T) {
	out := run(t, lines(
		"2", "3", "Шкаф у корпуса", "-1", "3,6", // Отрицательный расход зарядного шкафа
		"2", "1", "Шлем M", // шлем без расхода
		"3", "4", "5", "0",
	))
	assertContains(t, out,
		"Введите число от 0 до 1000",
		"[на балансе] Зарядный шкаф «Шкаф у корпуса», инв. № ВК-000001",
		"[на балансе] Шлем «Шлем M», инв. № ВК-000002",
		"3,60 кВт·ч",
		"Подходящего транспорта нет.",
	)
}

func TestEndOfInputStopsGracefully(t *testing.T) {
	// Ввод кончился посреди приёма самоката
	out := run(t, lines("1", "2", "Kugoo"))
	assertContains(t, out, "Простота для новичка")
}

func TestDemoFromMenu(t *testing.T) {
	out := run(t, lines("6", "0"))
	assertContains(t, out,
		"=== 0. Состояние парка до демо ===",
		"На балансе ничего нет.",
		"Подходящего транспорта нет.",
		"[принят]   Велосипед «Stels Navigator 300», инв. № ВК-000001",
		"[отклонён] Электросамокат «Xiaomi Pro 2»: неисправны тормоза",
		"повреждена рама; износ батареи: 40% при минимуме 70%",
		"объект уже на балансе: ВК-000002",
		"простота для новичка должна быть от 1 до 10",
		"суточный расход энергии должен быть числом от 0 до 1000 кВт·ч",
		"транспорт ставится на баланс только через техосмотр",
		"1,80 кВт·ч", // транспорт
		"4,80 кВт·ч", // вещи
		"6,60 кВт·ч", // всего
		"ВК-000008  Зарядный шкаф",
	)
	if strings.Contains(out, "Eltreco XT 800      4") {
		t.Error("транспорт с простотой 4 попал в список для новичков")
	}
}

// Отработка token too long
func TestVeryLongLineIsAccepted(t *testing.T) {
	long := strings.Repeat("а", 70_000)
	out := run(t, lines("1", "1", long, "7", "д", "д", "0"))
	assertContains(t, out, "[принят]   Велосипед")
}

// Отработка +Inf
func TestAbsurdEnergyIsReAsked(t *testing.T) {
	out := run(t, lines("2", "2", "Док", "1e308", "1,2", "3", "0"))
	assertContains(t, out, "Введите число от 0 до 1000", "1,20 кВт·ч")
}
