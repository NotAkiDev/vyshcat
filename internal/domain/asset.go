// Package domain описывает, из чего состоит парк, что можно поставить на баланс

package domain

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// MaxDailyKWh — верхняя граница суточного расхода одного устройства
const MaxDailyKWh = 1000

// Ошибки, которые возвращают конструкторы и AssignInventoryNumber
var (
	ErrEmptyName       = errors.New("наименование не может быть пустым")
	ErrInvalidNumber   = errors.New("инвентарный номер должен быть положительным")
	ErrAlreadyNumbered = errors.New("инвентарный номер уже присвоен")
	ErrInvalidEnergy   = fmt.Errorf("суточный расход энергии должен быть числом от 0 до %d кВт·ч", MaxDailyKWh)
)

// InventoryNumber — инвентарный номер
type InventoryNumber int

// String печатает стандартизированный инвентарный номер
func (n InventoryNumber) String() string {
	if n <= 0 {
		return "без номера"
	}
	return fmt.Sprintf("ВК-%06d", int(n))
}

// Inventoried реализует всё, что числится на балансе
type Inventoried interface {
	Name() string
	// Category отдаёт вид объекта для людей, по типу шлема или велосипеда
	Category() string
	InventoryNumber() InventoryNumber
	// AssignInventoryNumber вызывает парк при постановке на баланс
	AssignInventoryNumber(n InventoryNumber) error
}

// EnergyConsumer реализуют только те, кто реально потребляет энергию
type EnergyConsumer interface {
	// DailyEnergyKWh возвращает расход за сутки в киловатт/час
	DailyEnergyKWh() float64
}

// asset хранит то, что есть у любого объекта - имя и номер
type asset struct {
	name   string
	number InventoryNumber
}

// validName обрезает пробелы по краям и не пропускает пустое имя
func validName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", ErrEmptyName
	}
	return name, nil
}

func (a *asset) Name() string { return a.name }

func (a *asset) InventoryNumber() InventoryNumber { return a.number }

// AssignInventoryNumber не даёт перенумеровать объект
func (a *asset) AssignInventoryNumber(n InventoryNumber) error {
	if n <= 0 {
		return ErrInvalidNumber
	}
	if a.number != 0 {
		return fmt.Errorf("%w: %s", ErrAlreadyNumbered, a.number)
	}
	a.number = n
	return nil
}

// powered встраивается в типы, у которых есть суточный расход
type powered struct {
	dailyKWh float64
}

// validateEnergy отсекает отрицательный расход
func validateEnergy(dailyKWh float64) error {
	if math.IsNaN(dailyKWh) || dailyKWh < 0 || dailyKWh > MaxDailyKWh {
		return ErrInvalidEnergy
	}
	return nil
}

func (p *powered) DailyEnergyKWh() float64 { return p.dailyKWh }
