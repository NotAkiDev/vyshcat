// Package di собирает приложение в контейнере go.uber.org/dig
package di

import (
	"fmt"
	"io"

	"go.uber.org/dig"

	"vyshkat/internal/catalog"
	"vyshkat/internal/console"
	"vyshkat/internal/fleet"
	"vyshkat/internal/inspection"
)

// Config - всё, что можно настроить снаружи
type Config struct {
	Inspection inspection.Rules
	Beginners  fleet.BeginnerRule
	IO         console.IO
}

// DefaultConfig - стандартные пороги и переданные потоки ввода-вывода
func DefaultConfig(in io.Reader, out io.Writer) Config {
	return Config{
		Inspection: inspection.DefaultRules(),
		Beginners:  fleet.DefaultBeginnerRule(),
		IO:         console.IO{In: in, Out: out},
	}
}

// NewContainer Регистрирует конструкторы
func NewContainer(cfg Config) (*dig.Container, error) {
	c := dig.New()
	providers := []struct {
		constructor any
		opts        []dig.ProvideOption
	}{
		{constructor: func() inspection.Rules { return cfg.Inspection }},
		{constructor: func() fleet.BeginnerRule { return cfg.Beginners }},
		{constructor: func() console.IO { return cfg.IO }},
		{constructor: inspection.NewServiceCenter, opts: []dig.ProvideOption{dig.As(new(fleet.Inspector))}},
		{constructor: fleet.New},
		{constructor: catalog.Default},
		{constructor: console.New},
	}
	for _, p := range providers {
		if err := c.Provide(p.constructor, p.opts...); err != nil {
			return nil, fmt.Errorf("di: %w", err)
		}
	}
	return c, nil
}
